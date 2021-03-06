package typecheck

import (
	"java-mini-ls-go/javaparser"
	"java-mini-ls-go/parse"
	"java-mini-ls-go/parse/loc"
	"java-mini-ls-go/parse/typ"
	"java-mini-ls-go/util"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// GatherTypes traverses the given parse tree and gathers all class, method, field, etc. declarations.
// TODO doesn't get visibility of any types.
func GatherTypes(fileURI string, fileVersion int, tree antlr.Tree, builtins *typ.TypeMap) (*typ.TypeMap, *DefinitionsUsagesLookup) {
	userTypes := typ.NewTypeMap()
	defUsages := NewDefinitionsUsagesLookup()

	GatherTypesFirstPass(fileURI, fileVersion, tree, builtins, userTypes, defUsages)
	GatherTypesSecondPass(fileURI, fileVersion, tree, builtins, userTypes, defUsages)

	return userTypes, defUsages
}

func GatherTypesFirstPass(fileURI string, fileVersion int, tree antlr.Tree, builtins *typ.TypeMap, userTypes *typ.TypeMap, defUsages *DefinitionsUsagesLookup) {
	visitor := newTypeGatherer(fileURI, fileVersion, builtins, userTypes, defUsages)
	antlr.ParseTreeWalkerDefault.Walk(visitor, tree)
}

func GatherTypesSecondPass(fileURI string, fileVersion int, tree antlr.Tree, builtins *typ.TypeMap, userTypes *typ.TypeMap, defUsages *DefinitionsUsagesLookup) {
	visitor := newTypeGatherer(fileURI, fileVersion, builtins, userTypes, defUsages)
	visitor.setSecondPass()
	antlr.ParseTreeWalkerDefault.Walk(visitor, tree)
}

type formalParametersCtx interface {
	Identifier() javaparser.IIdentifierContext
	FormalParameters() javaparser.IFormalParametersContext
}

type methodCtx interface {
	formalParametersCtx
	TypeTypeOrVoid() javaparser.ITypeTypeOrVoidContext
}

type typeGatherer struct {
	javaparser.BaseJavaParserListener
	scopeTracker          *parse.ScopeTracker
	builtins              *typ.TypeMap
	userTypes             *typ.TypeMap
	defUsages             *DefinitionsUsagesLookup
	currFileURI           string
	currFileVersion       int
	currPackageName       string
	isFirstPass           bool
	currentMemberIsStatic bool
}

func newTypeGatherer(fileURI string, fileVersion int, builtins *typ.TypeMap, userTypes *typ.TypeMap, defUsages *DefinitionsUsagesLookup) *typeGatherer {
	return &typeGatherer{
		BaseJavaParserListener: javaparser.BaseJavaParserListener{},
		scopeTracker:           parse.NewScopeTracker(),
		builtins:               builtins,
		userTypes:              userTypes,
		defUsages:              defUsages,
		currFileURI:            fileURI,
		currFileVersion:        fileVersion,
		currPackageName:        "",
		isFirstPass:            true,
		currentMemberIsStatic:  false,
	}
}

func (tg *typeGatherer) makeCodeLocation(bounds loc.Bounds) loc.CodeLocation {
	return loc.CodeLocation{
		FileUri: tg.currFileURI,
		Version: tg.currFileVersion,
		Loc:     bounds,
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
		tg.addNewTypeFromScope(newScope, typ.JavaTypeClass)
	case parse.ScopeTypeInterface:
		tg.addNewTypeFromScope(newScope, typ.JavaTypeInterface)
	case parse.ScopeTypeEnum:
		tg.addNewTypeFromScope(newScope, typ.JavaTypeEnum)
	case parse.ScopeTypeAnnotationType:
		tg.addNewTypeFromScope(newScope, typ.JavaTypeAnnotation)
	case parse.ScopeTypeRecord:
		tg.addNewTypeFromScope(newScope, typ.JavaTypeRecord)
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
	if tg.isFirstPass {
		return
	}

	currTypeName := tg.scopeTracker.ScopeStack.Top().Name
	currType := tg.userTypes.Get(currTypeName)

	fieldTypeName := ctx.TypeType().GetText()
	fieldType := tg.lookupType(fieldTypeName)

	varDeclsI := ctx.VariableDeclarators()
	if varDeclsI != nil {
		varDecls := varDeclsI.(*javaparser.VariableDeclaratorsContext)
		for _, varDecl := range varDecls.AllVariableDeclarator() {
			ident := varDecl.(*javaparser.VariableDeclaratorContext).VariableDeclaratorId().(*javaparser.VariableDeclaratorIdContext).Identifier()
			fieldName := ident.GetText()
			bounds := loc.ParserRuleContextToBounds(ident)
			defLocation := tg.makeCodeLocation(bounds)

			field := &typ.JavaField{
				Name:       fieldName,
				Type:       fieldType,
				ParentType: currType,
				Definition: &defLocation,
				Usages:     []loc.CodeLocation{},
				Visibility: 0,
				IsStatic:   tg.currentMemberIsStatic,
				// TODO real value for IsFinal
				IsFinal: false,
			}

			currType.Fields = append(currType.Fields, field)

			tg.defUsages.Add(defLocation, field, false)
		}
	}
}

func (tg *typeGatherer) addNewTypeFromScope(scope *parse.Scope, ttype typ.JavaTypeType) {
	location := tg.makeCodeLocation(scope.Bounds)
	newType := typ.NewJavaType(scope.Name, tg.currPackageName, typ.VisibilityPublic, ttype, &location)
	tg.userTypes.Add(newType)
	tg.defUsages.Add(location, newType, false)
}

func (tg *typeGatherer) checkScopeExtendsImplements(scope *parse.Scope, ctx antlr.ParserRuleContext) {
	existingType := tg.lookupType(scope.Name)
	existingType.Extends = tg.getExtendsTypes(ctx)
	existingType.Implements = tg.getImplementsTypes(ctx)
	// TODO add existingType.Permits if it's relevant (new java 17 feature I think)
}

func (tg *typeGatherer) getExtendsTypes(ctx antlr.ParserRuleContext) []*typ.JavaType {
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

	return util.Map(typeTypes, func(typeType *javaparser.TypeTypeContext) *typ.JavaType {
		extendsTypeName := typeType.ClassOrInterfaceType().GetText()
		return tg.lookupType(extendsTypeName)
	})
}

func (tg *typeGatherer) getImplementsTypes(ctx antlr.ParserRuleContext) []*typ.JavaType {
	tctx, ok := ctx.(*javaparser.ClassDeclarationContext)
	if !ok {
		return []*typ.JavaType{}
	}

	implementsI := tctx.ClassDeclarationImplements()
	if implementsI != nil {
		typeList := implementsI.(*javaparser.ClassDeclarationImplementsContext).TypeList().(*javaparser.TypeListContext)

		implTypes := []*typ.JavaType{}
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

	return []*typ.JavaType{}
}

func (tg *typeGatherer) addNewConstructorFromScope(ctx formalParametersCtx) {
	// The top is the current scope, so we use top minus 1 to get the enclosing class
	currTypeName := tg.scopeTracker.ScopeStack.TopMinus(1).Name
	currType := tg.userTypes.Get(currTypeName)

	location := tg.makeCodeLocation(loc.ParserRuleContextToBounds(ctx.Identifier()))
	newConstructor := &typ.JavaConstructor{
		ParentType: currType,
		Params:     tg.getArgsFromContext(ctx),
		Definition: &location,
		Usages:     []loc.CodeLocation{},
		Visibility: 0,
	}

	currType.Constructors = append(currType.Constructors, newConstructor)

	tg.defUsages.Add(location, newConstructor, false)
}

func (tg *typeGatherer) addNewMethodFromScope(scope *parse.Scope, ctx methodCtx) {
	// The top is the current scope, so we use top minus 1 to get the enclosing class
	currTypeName := tg.scopeTracker.ScopeStack.TopMinus(1).Name
	currType := tg.userTypes.Get(currTypeName)

	location := tg.makeCodeLocation(loc.ParserRuleContextToBounds(ctx.Identifier()))
	method := &typ.JavaMethod{
		Name:       scope.Name,
		ParentType: currType,
		ReturnType: nil,
		Params:     nil,
		Definition: &location,
		Usages:     []loc.CodeLocation{},
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

	tg.defUsages.Add(location, method, false)
}

func (tg *typeGatherer) getArgsFromContext(ctx formalParametersCtx) []*typ.JavaParameter {
	args := make([]*typ.JavaParameter, 0)

	argsCtx := ctx.FormalParameters().(*javaparser.FormalParametersContext)

	receiverParameterCtx := argsCtx.ReceiverParameter()
	if receiverParameterCtx != nil {
		receiverParameter := receiverParameterCtx.(*javaparser.ReceiverParameterContext)
		arg := &typ.JavaParameter{
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
			arg := &typ.JavaParameter{
				Name:      argCtx.VariableDeclaratorId().GetText(),
				Type:      tg.lookupType(argCtx.TypeType().GetText()),
				IsVarargs: false,
			}
			args = append(args, arg)
		}

		lastParamI := paramList.LastFormalParameter()
		if lastParamI != nil {
			lastParam := lastParamI.(*javaparser.LastFormalParameterContext)
			arg := &typ.JavaParameter{
				Name:      lastParam.VariableDeclaratorId().GetText(),
				Type:      tg.lookupType(lastParam.TypeType().GetText()),
				IsVarargs: lastParam.ELLIPSIS() != nil,
			}
			args = append(args, arg)
		}
	}

	return args
}

func (tg *typeGatherer) lookupType(typeName string) *typ.JavaType {
	userType := tg.userTypes.Get(typeName)
	if userType != nil {
		return userType
	}

	return tg.builtins.Get(typeName)
}
