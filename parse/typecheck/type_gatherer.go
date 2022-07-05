package typecheck

import (
	"java-mini-ls-go/javaparser"
	"java-mini-ls-go/parse"
	"java-mini-ls-go/util"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// GatherTypes traverses the given parse tree and gathers all class, method, field, etc. declarations.
// TODO doesn't get visibility of any types.
func GatherTypes(fileURI string, tree antlr.Tree, builtins parse.TypeMap) parse.TypeMap {
	visitor := newTypeGatherer(fileURI, builtins)

	// First pass: just get types (no fields/methods yet, since those will reference the types)
	antlr.ParseTreeWalkerDefault.Walk(visitor, tree)

	// Second pass: populate fields/methods on every type
	visitor.setSecondPass()
	antlr.ParseTreeWalkerDefault.Walk(visitor, tree)

	return visitor.types
}

type formalParametersCtx interface {
	FormalParameters() javaparser.IFormalParametersContext
}

type methodCtx interface {
	formalParametersCtx
	TypeTypeOrVoid() javaparser.ITypeTypeOrVoidContext
}

type typeGatherer struct {
	javaparser.BaseJavaParserListener
	scopeTracker          *parse.ScopeTracker
	builtins              parse.TypeMap
	types                 parse.TypeMap
	defUsages             *DefinitionsUsagesLookup
	currFileURI           string
	currPackageName       string
	isFirstPass           bool
	currentMemberIsStatic bool
}

func newTypeGatherer(fileURI string, builtins parse.TypeMap) *typeGatherer {
	return &typeGatherer{
		BaseJavaParserListener: javaparser.BaseJavaParserListener{},
		scopeTracker:           parse.NewScopeTracker(),
		builtins:               builtins,
		types:                  make(parse.TypeMap),
		defUsages:              NewDefinitionsUsagesLookup(),
		currFileURI:            fileURI,
		currPackageName:        "",
		isFirstPass:            true,
		currentMemberIsStatic:  false,
	}
}

func (tg *typeGatherer) setSecondPass() {
	// Reset state for second pass
	tg.isFirstPass = false
	tg.currPackageName = ""
}

func (tg *typeGatherer) EnterEveryRule(ctx antlr.ParserRuleContext) {
	newScope := tg.scopeTracker.CheckEnterScope(ctx)
	if newScope != nil {
		if tg.isFirstPass {
			tg.handleNewScopeFirstPass(newScope, ctx)
		} else {
			tg.handleNewScopeSecondPass(newScope, ctx)
		}
	}
}

func (tg *typeGatherer) handleNewScopeFirstPass(newScope *parse.Scope, _ antlr.ParserRuleContext) {
	switch newScope.Type {
	case parse.ScopeTypeClass:
		tg.addNewTypeFromScope(newScope, parse.JavaTypeClass)
	case parse.ScopeTypeInterface:
		tg.addNewTypeFromScope(newScope, parse.JavaTypeInterface)
	case parse.ScopeTypeEnum:
		tg.addNewTypeFromScope(newScope, parse.JavaTypeEnum)
	case parse.ScopeTypeAnnotationType:
		tg.addNewTypeFromScope(newScope, parse.JavaTypeAnnotation)
	case parse.ScopeTypeRecord:
		tg.addNewTypeFromScope(newScope, parse.JavaTypeRecord)
	}
}

func (tg *typeGatherer) handleNewScopeSecondPass(scope *parse.Scope, ctx antlr.ParserRuleContext) {
	switch scope.Type {
	case parse.ScopeTypeClass:
		tg.checkScopeExtendsImplements(scope, ctx)
	case parse.ScopeTypeInterface:
		tg.checkScopeExtendsImplements(scope, ctx)

	case parse.ScopeTypeConstructor:
		fallthrough
	case parse.ScopeTypeGenericConstructor:
		tg.addNewConstructorFromScope(ctx.(formalParametersCtx))

	case parse.ScopeTypeMethod:
		fallthrough
	case parse.ScopeTypeGenericMethod:
		fallthrough
	case parse.ScopeTypeInterfaceMethod:
		fallthrough
	case parse.ScopeTypeGenericInterfaceMethod:
		tg.addNewMethodFromScope(scope, ctx.(methodCtx))
	}
}

func (tg *typeGatherer) ExitEveryRule(ctx antlr.ParserRuleContext) {
	_ = tg.scopeTracker.CheckExitScope(ctx)
}

// EnterPackageDeclaration is called when production packageDeclaration is entered.
func (tg *typeGatherer) EnterPackageDeclaration(ctx *javaparser.PackageDeclarationContext) {
	tg.currPackageName = ctx.QualifiedName().GetText()
}

// EnterClassBodyDeclaration is called when production classBodyDeclaration is entered.
func (tg *typeGatherer) EnterClassBodyDeclaration(ctx *javaparser.ClassBodyDeclarationContext) {
	if ctx.STATIC() != nil {
		// TODO is it possible for class body declarations to be nested? In that case this would
		// need to be a stack instead of a single bool
		tg.currentMemberIsStatic = true
	}
}

// ExitClassBodyDeclaration is called when production classBodyDeclaration is exited.
func (tg *typeGatherer) ExitClassBodyDeclaration(ctx *javaparser.ClassBodyDeclarationContext) {
	if ctx.STATIC() != nil {
		// If this is true, we set currentMemberIsStatic to true on the way in.
		// So we want to make sure to set it to false on the way out.
		tg.currentMemberIsStatic = false
	}
}

// EnterFieldDeclaration is called when production fieldDeclaration is entered.
func (tg *typeGatherer) EnterFieldDeclaration(ctx *javaparser.FieldDeclarationContext) {
	currTypeName := tg.scopeTracker.ScopeStack.Top().Name
	currType := tg.types[currTypeName]

	fieldTypeName := ctx.TypeType().GetText()
	fieldType := tg.lookupType(fieldTypeName)

	varDeclsI := ctx.VariableDeclarators()
	if varDeclsI != nil {
		varDecls := varDeclsI.(*javaparser.VariableDeclaratorsContext)
		for _, varDecl := range varDecls.AllVariableDeclarator() {
			field := &parse.JavaField{
				Name:       varDecl.GetText(),
				Type:       fieldType,
				ParentType: currType,
				Definition: nil,
				Usages:     []parse.CodeLocation{},
				Visibility: 0,
				IsStatic:   tg.currentMemberIsStatic,
				// TODO real value for IsFinal
				IsFinal: false,
			}

			currType.Fields = append(currType.Fields, field)
		}
	}
}

func (tg *typeGatherer) addNewTypeFromScope(scope *parse.Scope, ttype parse.JavaTypeType) {
	newType := parse.NewJavaType(scope.Name, tg.currPackageName, parse.VisibilityPublic, ttype)
	tg.types[scope.Name] = newType
	tg.defUsages.NewSymbol(scope.Bounds, newType)
}

func (tg *typeGatherer) checkScopeExtendsImplements(scope *parse.Scope, ctx antlr.ParserRuleContext) {
	existingType := tg.lookupType(scope.Name)
	existingType.Extends = tg.getExtendsTypes(ctx)
	existingType.Implements = tg.getImplementsTypes(ctx)
	// TODO add existingType.Permits if it's relevant (new java 17 feature I think)
}

func (tg *typeGatherer) getExtendsTypes(ctx antlr.ParserRuleContext) []*parse.JavaType {
	typeTypes := []*javaparser.TypeTypeContext{}

	switch tctx := ctx.(type) {
	case *javaparser.ClassDeclarationContext:
		extendsI := tctx.ClassDeclarationExtends()
		if extendsI != nil {
			extends := extendsI.(*javaparser.ClassDeclarationExtendsContext)
			typeTypes = []*javaparser.TypeTypeContext{
				extends.TypeType().(*javaparser.TypeTypeContext),
			}
		}
	case *javaparser.InterfaceDeclarationContext:
		extendsI := tctx.InterfaceDeclarationExtends()
		if extendsI != nil {
			extends := extendsI.(*javaparser.InterfaceDeclarationExtendsContext)
			extendsTypeList := extends.TypeList().(*javaparser.TypeListContext)
			allTypeTypes := extendsTypeList.AllTypeType()
			for _, tt := range allTypeTypes {
				if tt != nil {
					typeTypes = append(typeTypes, tt.(*javaparser.TypeTypeContext))
				}
			}
		}
	}

	return util.Map(typeTypes, func(typeType *javaparser.TypeTypeContext) *parse.JavaType {
		extendsTypeName := typeType.ClassOrInterfaceType().GetText()
		return tg.lookupType(extendsTypeName)
	})
}

func (tg *typeGatherer) getImplementsTypes(ctx antlr.ParserRuleContext) []*parse.JavaType {
	tctx, ok := ctx.(*javaparser.ClassDeclarationContext)
	if !ok {
		return []*parse.JavaType{}
	}

	implementsI := tctx.ClassDeclarationImplements()
	if implementsI != nil {
		typeList := implementsI.(*javaparser.ClassDeclarationImplementsContext).TypeList().(*javaparser.TypeListContext)

		implTypes := []*parse.JavaType{}
		allTypeTypes := typeList.AllTypeType()
		for _, tt := range allTypeTypes {
			if tt != nil {
				typeType := tt.(*javaparser.TypeTypeContext)
				extendsTypeName := typeType.ClassOrInterfaceType().GetText()
				implTypes = append(implTypes, tg.lookupType(extendsTypeName))
			}
		}
		return implTypes
	}

	return []*parse.JavaType{}
}

func (tg *typeGatherer) addNewConstructorFromScope(ctx formalParametersCtx) {
	// The top is the current scope, so we use top minus 1 to get the enclosing class
	currTypeName := tg.scopeTracker.ScopeStack.TopMinus(1).Name
	currType := tg.types[currTypeName]

	newConstructor := &parse.JavaConstructor{
		ParentType: currType,
		Params:     tg.getArgsFromContext(ctx),
		Definition: &parse.CodeLocation{
			FileUri: tg.currFileURI,
			Loc:     parse.ParserRuleContextToBounds(ctx.(antlr.ParserRuleContext)),
		},
		Usages:     []parse.CodeLocation{},
		Visibility: 0,
	}

	currType.Constructors = append(currType.Constructors, newConstructor)
}

func (tg *typeGatherer) addNewMethodFromScope(scope *parse.Scope, ctx methodCtx) {
	// The top is the current scope, so we use top minus 1 to get the enclosing class
	currTypeName := tg.scopeTracker.ScopeStack.TopMinus(1).Name
	currType := tg.types[currTypeName]

	method := &parse.JavaMethod{
		Name:       scope.Name,
		ParentType: currType,
		ReturnType: nil,
		Params:     nil,
		Definition: &parse.CodeLocation{
			FileUri: tg.currFileURI,
			Loc:     parse.ParserRuleContextToBounds(ctx.(antlr.ParserRuleContext)),
		},
		Usages:     []parse.CodeLocation{},
		Visibility: 0,
		IsStatic:   false,
	}

	returnType := ctx.TypeTypeOrVoid().GetText()
	if returnType != "void" {
		method.ReturnType = tg.lookupType(returnType)
	}

	method.Params = tg.getArgsFromContext(ctx)
	method.IsStatic = tg.currentMemberIsStatic

	currType.Methods = append(currType.Methods, method)
}

func (tg *typeGatherer) getArgsFromContext(ctx formalParametersCtx) []*parse.JavaParameter {
	args := make([]*parse.JavaParameter, 0)

	argsCtx := ctx.FormalParameters().(*javaparser.FormalParametersContext)

	receiverParameterCtx := argsCtx.ReceiverParameter()
	if receiverParameterCtx != nil {
		receiverParameter := receiverParameterCtx.(*javaparser.ReceiverParameterContext)
		arg := &parse.JavaParameter{
			Name:      "this",
			Type:      tg.lookupType(receiverParameter.TypeType().GetText()),
			IsVarargs: false,
		}
		args = append(args, arg)
	}

	paramListI := argsCtx.FormalParameterList()
	if paramListI != nil {
		// sooo much interface casting, idk why they do this
		paramList := paramListI.(*javaparser.FormalParameterListContext)
		for _, argICtx := range paramList.AllFormalParameter() {
			argCtx := argICtx.(*javaparser.FormalParameterContext)
			arg := &parse.JavaParameter{
				Name:      argCtx.VariableDeclaratorId().GetText(),
				Type:      tg.lookupType(argCtx.TypeType().GetText()),
				IsVarargs: false,
			}
			args = append(args, arg)
		}

		lastParamI := paramList.LastFormalParameter()
		if lastParamI != nil {
			lastParam := lastParamI.(*javaparser.LastFormalParameterContext)
			arg := &parse.JavaParameter{
				Name:      lastParam.VariableDeclaratorId().GetText(),
				Type:      tg.lookupType(lastParam.TypeType().GetText()),
				IsVarargs: lastParam.ELLIPSIS() != nil,
			}
			args = append(args, arg)
		}
	}

	return args
}

func (tg *typeGatherer) lookupType(typeName string) *parse.JavaType {
	userType, ok := tg.types[typeName]
	if ok {
		return userType
	}

	return tg.builtins[typeName]
}
