package parse

import (
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"java-mini-ls-go/javaparser"
)

// GatherTypes traverses the given parse tree and gathers all class, method, field, etc. declarations.
// TODO doesn't get visibility of any types.
func GatherTypes(tree antlr.Tree, builtins TypeMap) TypeMap {
	visitor := &typeGatherer{
		scopeTracker: NewScopeTracker(),
		builtins:     builtins,
		types:        make(TypeMap),
		isFirstPass:  true,
	}

	// First pass: just get types (no fields/methods yet, since those will reference the types)
	antlr.ParseTreeWalkerDefault.Walk(visitor, tree)

	// Second pass: populate fields/methods on every type
	visitor.isFirstPass = false
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
	scopeTracker          *ScopeTracker
	builtins              TypeMap
	types                 TypeMap
	currPackageName       string
	isFirstPass           bool
	currentMemberIsStatic bool
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

func (tg *typeGatherer) handleNewScopeFirstPass(newScope *Scope, _ antlr.ParserRuleContext) {
	switch newScope.Type {
	case ScopeTypeClass:
		tg.addNewTypeFromScope(newScope, JavaTypeClass)
	case ScopeTypeInterface:
		tg.addNewTypeFromScope(newScope, JavaTypeInterface)
	case ScopeTypeEnum:
		tg.addNewTypeFromScope(newScope, JavaTypeEnum)
	case ScopeTypeAnnotationType:
		tg.addNewTypeFromScope(newScope, JavaTypeAnnotation)
	case ScopeTypeRecord:
		tg.addNewTypeFromScope(newScope, JavaTypeRecord)
	}
}

func (tg *typeGatherer) handleNewScopeSecondPass(scope *Scope, ctx antlr.ParserRuleContext) {
	switch scope.Type {
	case ScopeTypeClass:
		tg.checkScopeExtendsImplements(newScope, ctx)
	case ScopeTypeInterface:
		tg.checkScopeExtendsImplements(newScope, ctx)

	case ScopeTypeConstructor:
		fallthrough
	case ScopeTypeGenericConstructor:
		tg.addNewConstructorFromScope(scope, ctx.(formalParametersCtx))

	case ScopeTypeMethod:
		fallthrough
	case ScopeTypeGenericMethod:
		fallthrough
	case ScopeTypeInterfaceMethod:
		fallthrough
	case ScopeTypeGenericInterfaceMethod:
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
			field := &JavaField{
				Name:     varDecl.GetText(),
				Type:     fieldType,
				IsStatic: tg.currentMemberIsStatic,
			}

			currType.Fields[field.Name] = field
		}
	}
}

func (tg *typeGatherer) addNewTypeFromScope(scope *Scope, ttype JavaTypeType) {
	newType := &JavaType{
		Name:         scope.Name,
		Package:      tg.currPackageName,
		Constructors: make([]*JavaConstructor, 0),
		Fields:       make(map[string]*JavaField),
		Methods:      make(map[string]*JavaMethod),
		Type:         ttype,
	}

	tg.types[scope.Name] = newType
}

func (tg *typeGatherer) checkScopeExtendsImplements(scope *Scope, ctx *javaparser.ClassDeclarationContext) {
	existingType := tg.lookupType(scope.Name)

	extends := ctx.EXTENDS()
	if extends != nil {
		
	}

	existingType.Extends = tg.lookupType()
}

func (tg *typeGatherer) addNewConstructorFromScope(scope *Scope, ctx formalParametersCtx) {
	// The top is the current scope, so we use top minus 1 to get the enclosing class
	currTypeName := tg.scopeTracker.ScopeStack.TopMinus(1).Name
	currType := tg.types[currTypeName]

	newConstructor := &JavaConstructor{
		Arguments: tg.getArgsFromContext(ctx),
	}

	currType.Constructors = append(currType.Constructors, newConstructor)
}

func (tg *typeGatherer) addNewMethodFromScope(scope *Scope, ctx methodCtx) {
	// The top is the current scope, so we use top minus 1 to get the enclosing class
	currTypeName := tg.scopeTracker.ScopeStack.TopMinus(1).Name
	currType := tg.types[currTypeName]

	method := &JavaMethod{
		Name:       scope.Name,
		ReturnType: nil,
		Params:     nil,
		IsStatic:   false,
	}

	returnType := ctx.TypeTypeOrVoid().GetText()
	if returnType != "void" {
		method.ReturnType = tg.lookupType(returnType)
	}

	method.Params = tg.getArgsFromContext(ctx)
	method.IsStatic = tg.currentMemberIsStatic

	currType.Methods[method.Name] = method
}

func (tg *typeGatherer) getArgsFromContext(ctx formalParametersCtx) []*JavaParameter {
	args := make([]*JavaParameter, 0)

	argsCtx := ctx.FormalParameters().(*javaparser.FormalParametersContext)

	receiverParameterCtx := argsCtx.ReceiverParameter()
	if receiverParameterCtx != nil {
		receiverParameter := receiverParameterCtx.(*javaparser.ReceiverParameterContext)
		arg := &JavaParameter{
			Name: "this",
			Type: tg.lookupType(receiverParameter.TypeType().GetText()),
		}
		args = append(args, arg)
	}

	paramListI := argsCtx.FormalParameterList()
	if paramListI != nil {
		// sooo much interface casting, idk why they do this
		paramList := paramListI.(*javaparser.FormalParameterListContext)
		for _, argICtx := range paramList.AllFormalParameter() {
			argCtx := argICtx.(*javaparser.FormalParameterContext)
			arg := &JavaParameter{
				Name: argCtx.VariableDeclaratorId().GetText(),
				Type: tg.lookupType(argCtx.TypeType().GetText()),
			}
			args = append(args, arg)
		}

		lastParamI := paramList.LastFormalParameter()
		if lastParamI != nil {
			lastParam := lastParamI.(*javaparser.LastFormalParameterContext)
			arg := &JavaParameter{
				Name:      lastParam.VariableDeclaratorId().GetText(),
				Type:      tg.lookupType(lastParam.TypeType().GetText()),
				IsVarargs: lastParam.ELLIPSIS() != nil,
			}
			args = append(args, arg)
		}
	}

	return args
}

func (tg *typeGatherer) lookupType(typeName string) *JavaType {
	userType, ok := tg.types[typeName]
	if ok {
		return userType
	}

	return tg.builtins[typeName]
}
