package typecheck

import (
	"fmt"
	"java-mini-ls-go/javaparser"
	"java-mini-ls-go/parse"
	"java-mini-ls-go/parse/loc"
	"java-mini-ls-go/parse/typ"
	"java-mini-ls-go/util"
	"math"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

type ExprType int

const (
	ExprTypeUnset   ExprType = iota
	ExprTypeDotExpr ExprType = iota

	// Other expression types, not used yet
	//ExprTypeExpression     ExprType = iota
	//ExprTypePrimary        ExprType = iota
	//ExprTypeIndexExpr      ExprType = iota
	//ExprTypeMethodCallExpr ExprType = iota
	//ExprTypeNewExpr        ExprType = iota
	//ExprTypeCastExpr       ExprType = iota
	//ExprTypeUnaryExpr      ExprType = iota
	//ExprTypeBinaryExpr     ExprType = iota
	//ExprTypeTernaryExpr    ExprType = iota
	//ExprTypeLambdaExpr     ExprType = iota
	//ExprTypeSwitchExpr     ExprType = iota
	//ExprTypeExprOther      ExprType = iota
)

//func expressionTypeForCtx(ctx *javaparser.ExpressionContext) ScopeType {
//	if ctx.Primary() != nil {
//		return ScopeTypePrimary
//	}
//	if ctx.GetDotop() != nil {
//		return ScopeTypeDotExpr
//	}
//	if ctx.GetIndexop() != nil {
//		return ScopeTypeIndexExpr
//	}
//	if ctx.MethodCall() != nil {
//		return ScopeTypeMethodCallExpr
//	}
//	if ctx.NEW() != nil {
//		return ScopeTypeNewExpr
//	}
//	if ctx.CastExpr() != nil {
//		return ScopeTypeCastExpr
//	}
//	if ctx.GetPostfix() != nil || ctx.GetPrefix() != nil {
//		return ScopeTypeUnaryExpr
//	}
//	if ctx.GetBop() != nil {
//		return ScopeTypeBinaryExpr
//	}
//	if ctx.GetTern() != nil {
//		return ScopeTypeTernaryExpr
//	}
//	if ctx.LambdaExpression() != nil {
//		return ScopeTypeLambdaExpr
//	}
//	if ctx.SwitchExpression() != nil {
//		return ScopeTypeSwitchExpr
//	}
//	return ScopeTypeExprOther
//}

type TypeError struct {
	Loc     loc.Bounds
	Message string
}

func (te *TypeError) ToDiagnostic() protocol.Diagnostic {
	return protocol.Diagnostic{
		Range:              loc.BoundsToRange(te.Loc),
		Severity:           protocol.DiagnosticSeverityError,
		Code:               nil,
		CodeDescription:    nil,
		Source:             "java-mini-ls",
		Message:            te.Message,
		Tags:               []protocol.DiagnosticTag{},
		RelatedInformation: []protocol.DiagnosticRelatedInformation{},
		Data:               nil,
	}
}

//goland:noinspection GoNameStartsWithPackageName
type TypeCheckResult struct {
	TypeErrors      []TypeError
	DefUsagesLookup *DefinitionsUsagesLookup
	RootScope       *TypeCheckingScope
}

// GatherAndCheckTypes is the entrypoint for type-related analysis involving one file. First calls GatherTypes and then CheckTypes,
// helpfully stringing the return values of the one into the ones that are necessary for the other.
func GatherAndCheckTypes(logger *zap.Logger, fileURI string, fileVersion int, tree antlr.Tree, builtins *typ.TypeMap) TypeCheckResult {
	types, defUsages := GatherTypes(fileURI, fileVersion, tree, builtins)
	return CheckTypes(logger, fileURI, fileVersion, tree, types, builtins, defUsages)
}

// CheckTypes traverses the given parse tree and performs type checking in all applicable
// places. e.g. expressions, return statements, function calls, etc.
func CheckTypes(logger *zap.Logger, fileURI string, fileVersion int, tree antlr.Tree, userTypes *typ.TypeMap, builtins *typ.TypeMap, defUsages *DefinitionsUsagesLookup) TypeCheckResult {
	visitor := newTypeChecker(logger, fileURI, fileVersion, builtins, userTypes, defUsages)
	antlr.ParseTreeWalkerDefault.Walk(visitor, tree)

	return TypeCheckResult{
		TypeErrors:      visitor.errors,
		DefUsagesLookup: defUsages,
		RootScope:       visitor.rootScope,
	}
}

type typedDeclarationCtx interface {
	TypeType() javaparser.ITypeTypeContext
	VariableDeclarators() javaparser.IVariableDeclaratorsContext
}

type typedExpression struct {
	loc   loc.Bounds
	ttype *typ.JavaType

	// Normally the above fields are all that's necessary. But sometimes we'll push a placeholder value
	// that doesn't have the above fields, and uses a scope type instead.
	placeholderExprType ExprType
}

func (te typedExpression) String() string {
	return fmt.Sprintf("loc=%v type=%v", te.loc, te.ttype)
}

type typeChecker struct {
	javaparser.BaseJavaParserListener
	logger          *zap.Logger
	currFileURI     string
	currFileVersion int
	userTypes       *typ.TypeMap
	builtins        *typ.TypeMap
	errors          []TypeError
	scopeTracker    *parse.ScopeTracker
	rootScope       *TypeCheckingScope
	currentScope    *TypeCheckingScope
	defUsages       *DefinitionsUsagesLookup

	// A stack used to keep track of the types of various expressions.
	// For example, in the binary expression `9 + 10`:
	// first, 9 will get evaluated, pushing the type "int" onto the stack
	// Then, 10 will get evaluated, pushing another "int" onto the stack
	// Then, the + binary expression gets evaluated, popping off the 2 last items
	// from the stack and checking their types.
	expressionStack util.Stack[typedExpression]
}

func newTypeChecker(logger *zap.Logger, fileURI string, fileVersion int, builtins *typ.TypeMap, userTypes *typ.TypeMap, defUsages *DefinitionsUsagesLookup) *typeChecker {
	rootScope := newTypeCheckingScope(
		nil,
		nil,
		// Bounds representing the entire file
		loc.Bounds{
			Start: loc.FileLocation{
				Line:      0,
				Character: 0,
			},
			End: loc.FileLocation{
				Line:      math.MaxInt,
				Character: math.MaxInt,
			},
		},
	)

	return &typeChecker{
		BaseJavaParserListener: javaparser.BaseJavaParserListener{},
		logger:                 logger.Named("typeChecker"),
		currFileURI:            fileURI,
		currFileVersion:        fileVersion,
		userTypes:              userTypes,
		builtins:               builtins,
		errors:                 make([]TypeError, 0),
		scopeTracker:           parse.NewScopeTracker(),
		rootScope:              rootScope,
		currentScope:           rootScope,
		defUsages:              defUsages,
		expressionStack:        util.NewStack[typedExpression](),
	}
}

func (tc *typeChecker) makeCodeLocation(bounds loc.Bounds) loc.CodeLocation {
	return loc.CodeLocation{
		FileUri: tc.currFileURI,
		Version: tc.currFileVersion,
		Loc:     bounds,
	}
}

func (tc *typeChecker) addError(err TypeError) {
	tc.errors = append(tc.errors, err)
}

func (tc *typeChecker) lookupType(typeName string) *typ.JavaType {
	userType := tc.userTypes.Get(typeName)
	if userType != nil {
		return userType
	}

	// will return nil if not found
	return tc.builtins.Get(typeName)
}

func (tc *typeChecker) lookupOrCreateType(typeName string) *typ.JavaType {
	found := tc.lookupType(typeName)
	if found != nil {
		return found
	}

	// Type doesn't exist, create it
	fmt.Println("Creating built-in type: ", typeName)
	jtype := typ.NewJavaType(typeName, "", typ.VisibilityPublic, typ.JavaTypeClass, nil)
	tc.builtins.Add(jtype)

	return jtype
}

func (tc *typeChecker) pushExprType(ttype *typ.JavaType, bounds loc.Bounds) {
	tc.expressionStack.Push(typedExpression{
		loc:                 bounds,
		ttype:               ttype,
		placeholderExprType: ExprTypeUnset,
	})
}

func (tc *typeChecker) pushExprTypeName(typeName string, bounds loc.Bounds) {
	tc.pushExprType(tc.lookupOrCreateType(typeName), bounds)
}

func (tc *typeChecker) pushAnyType(bounds loc.Bounds) {
	tc.pushExprTypeName("any", bounds)
}

func (tc *typeChecker) pushPlaceholder(exprType ExprType) {
	tc.expressionStack.Push(typedExpression{
		loc:                 loc.Bounds{}, //nolint:exhaustruct
		ttype:               nil,
		placeholderExprType: exprType,
	})
}

func (tc *typeChecker) insideExpressionType(exprType ExprType) bool {
	for _, typedExpr := range tc.expressionStack.Values() {
		if typedExpr.placeholderExprType == exprType {
			return true
		}
	}
	return false
}

// Pop everything off the expression stack until we get to a particular type of placeholder
func (tc *typeChecker) popUntilPlaceholderType(exprType ExprType) []typedExpression {
	ret := []typedExpression{}
	for {
		popped := tc.expressionStack.Pop()
		if popped.placeholderExprType == exprType {
			break
		}
		ret = append(ret, popped)
	}
	return ret
}

// checkAndAddVariable adds a local variable, while first checking whether the local
// is already defined, and if so, adding an error.
func (tc *typeChecker) checkAndAddVariable(name string, ttype *typ.JavaType, bounds loc.Bounds, scopeType string) {
	topScope := tc.currentScope
	if _, ok := topScope.Locals[name]; ok {
		currMethodName := tc.scopeTracker.ScopeStack.Top().Name
		tc.addError(TypeError{
			Loc:     bounds,
			Message: fmt.Sprintf("Variable %s is already defined in %s %s", name, scopeType, currMethodName),
		})
	}

	enclosingMethod, ok := topScope.Symbol.(*typ.JavaMethod)
	if ok {
		// We're inside a method, so it's a local
		local := typ.NewJavaLocal(name, ttype, enclosingMethod, tc.makeCodeLocation(bounds))
		topScope.addLocal(local)
		tc.defUsages.Add(tc.makeCodeLocation(bounds), local, false)
	}
}

func (tc *typeChecker) getEnclosingType() *typ.JavaType {
	scopes := tc.scopeTracker.ScopeStack
	for i := scopes.Size() - 1; i >= 0; i-- {
		scope := scopes.TopMinus(i)
		if scope.Type.IsClassType() {
			return tc.lookupOrCreateType(scope.Name)
		}
	}
	return nil
}

func (tc *typeChecker) EnterEveryRule(ctx antlr.ParserRuleContext) {
	newScope := tc.scopeTracker.CheckEnterScope(ctx)
	if newScope != nil {
		bounds := loc.ParserRuleContextToBounds(ctx)
		symbolForScope := tc.getSymbolFromScope(newScope)
		typeScope := newTypeCheckingScope(symbolForScope, tc.currentScope, bounds)

		if newScope.Type.IsMethodType() {
			// Add method params to Locals.
			// First, look up method in types.
			enclosingType := tc.getEnclosingType()
			if enclosingType != nil {
				// TODO handle method overrides (same name)
				methodIdx := slices.IndexFunc(enclosingType.Methods, func(method *typ.JavaMethod) bool {
					return method.Name == newScope.Name
				})
				if methodIdx != -1 {
					method := enclosingType.Methods[methodIdx]

					for _, param := range method.Params {
						local := typ.NewJavaLocal(param.Name, param.Type, method, tc.makeCodeLocation(bounds))
						typeScope.addLocal(local)
					}
				}
			}
		}

		tc.currentScope = typeScope
	}
}

func (tc *typeChecker) getSymbolFromScope(scope *parse.Scope) typ.JavaSymbol {
	if scope.Type.IsClassType() {
		return tc.lookupOrCreateType(scope.Name)
	}

	// it's a method, should be inside a class
	if tc.currentScope.Symbol == nil {
		tc.logger.Error(fmt.Sprintf("can't get symbol from scope, tc.currentScope.Symbol is nil. Scope: %v", scope))
		return nil
	}

	enclosingType, ok := tc.currentScope.Symbol.(*typ.JavaType)
	if !ok {
		tc.logger.Error(fmt.Sprintf("can't get symbol from scope, enclosing type is %T. Scope: %v", tc.currentScope.Symbol, scope))
		return nil
	}

	// Look up method on enclosingType
	return enclosingType.LookupMember(scope.Name)
}

func (tc *typeChecker) ExitEveryRule(ctx antlr.ParserRuleContext) {
	oldScope := tc.scopeTracker.CheckExitScope(ctx)
	if oldScope != nil {
		tc.currentScope = tc.currentScope.Parent
	}
}

func (tc *typeChecker) ExitStatement(_ *javaparser.StatementContext) {
	// zero out the expression stack when we leave a statement
	tc.expressionStack.Clear()
}

func (tc *typeChecker) ExitBlockStatement(_ *javaparser.BlockStatementContext) {
	// zero out the expression stack when we leave a statement
	tc.expressionStack.Clear()
}

func (tc *typeChecker) ExitFieldDeclaration(ctx *javaparser.FieldDeclarationContext) {
	tc.handleTypedVariableDecl(ctx, loc.ParserRuleContextToBounds(ctx), false)
}

func (tc *typeChecker) ExitLocalVariableDeclaration(ctx *javaparser.LocalVariableDeclarationContext) {
	typedI := ctx.TypedLocalVarDecl()
	if typedI != nil {
		typed := typedI.(*javaparser.TypedLocalVarDeclContext)
		tc.handleTypedVariableDecl(typed, loc.ParserRuleContextToBounds(ctx), true)
	} else {
		untyped := ctx.UntypedLocalVarDecl().(*javaparser.UntypedLocalVarDeclContext)
		tc.handleUntypedLocalVariableDecl(untyped)
	}
}

// e.g. `String a = "hi"`
func (tc *typeChecker) handleTypedVariableDecl(ctx typedDeclarationCtx, bounds loc.Bounds, isLocal bool) {
	ttype := tc.lookupOrCreateType(ctx.TypeType().GetText())

	// There can be multiple variable declarators
	varDecls := ctx.VariableDeclarators().(*javaparser.VariableDeclaratorsContext).AllVariableDeclarator()
	for _, varDeclI := range varDecls {
		varDecl := varDeclI.(*javaparser.VariableDeclaratorContext)

		ident := varDecl.VariableDeclaratorId().(*javaparser.VariableDeclaratorIdContext).Identifier()
		varName := ident.GetText()

		var scopeType string
		if isLocal {
			scopeType = "method"
		} else {
			scopeType = "class"
		}

		// TODO fix bounds, the error message also red underlines the equals sign
		tc.checkAndAddVariable(varName, ttype, loc.ParserRuleContextToBounds(ident), scopeType)
	}

	// Make sure every value in the expression stack (which is the value of all the initializer expressions
	// for these local vars) coerces to the type declared.
	for !tc.expressionStack.Empty() {
		expr := tc.expressionStack.Pop()
		if !expr.ttype.CoercesTo(ttype) {
			tc.addError(TypeError{
				Loc:     expr.loc,
				Message: fmt.Sprintf("Type mismatch: cannot convert from %s to %s", expr.ttype.Name, ttype.Name),
			})
		}
	}
}

// e.g. `var a = "hi"`
func (tc *typeChecker) handleUntypedLocalVariableDecl(ctx *javaparser.UntypedLocalVarDeclContext) {
	// In order for type to be inferred, we must have already pushed the expression type
	ttype := tc.expressionStack.Pop().ttype

	tc.checkAndAddVariable(ctx.Identifier().GetText(), ttype, loc.ParserRuleContextToBounds(ctx.Identifier()), "method")
}

func (tc *typeChecker) ExitPrimary(ctx *javaparser.PrimaryContext) {
	literal := ctx.Literal()
	if literal != nil {
		tc.handleLiteral(literal.(*javaparser.LiteralContext))
	}

	ident := ctx.Identifier()
	if ident != nil {
		tc.handleIdentifier(ident.(*javaparser.IdentifierContext))
	}
}

func (tc *typeChecker) handleLiteral(ctx *javaparser.LiteralContext) {
	bounds := loc.ParserRuleContextToBounds(ctx)

	intLit := ctx.IntegerLiteral()
	if intLit != nil {
		typeName := "int"

		intLitTxt := intLit.GetText()
		lastChar := intLitTxt[len(intLitTxt)-1]
		if lastChar == 'l' || lastChar == 'L' {
			typeName = "long"
		}

		tc.pushExprTypeName(typeName, bounds)
		return
	}

	floatLit := ctx.FloatLiteral()
	if floatLit != nil {
		typeName := "double"

		floatLitTxt := floatLit.GetText()
		lastChar := floatLitTxt[len(floatLitTxt)-1]
		if lastChar == 'f' || lastChar == 'F' {
			typeName = "float"
		}

		tc.pushExprTypeName(typeName, bounds)
		return
	}

	charLit := ctx.CHAR_LITERAL()
	if charLit != nil {
		tc.pushExprTypeName("char", bounds)
		return
	}

	strLit := ctx.STRING_LITERAL()
	if strLit != nil {
		tc.pushExprTypeName("String", bounds)
		return
	}

	boolLit := ctx.BOOL_LITERAL()
	if boolLit != nil {
		tc.pushExprTypeName("boolean", bounds)
		return
	}

	nullLit := ctx.NULL_LITERAL()
	if nullLit != nil {
		tc.pushAnyType(bounds)
		return
	}

	textBlockLit := ctx.TEXT_BLOCK()
	if textBlockLit != nil {
		tc.pushExprTypeName("String", bounds)
		return
	}
}

func (tc *typeChecker) handleIdentifier(ctx *javaparser.IdentifierContext) {
	bounds := loc.ParserRuleContextToBounds(ctx)
	identName := ctx.GetText()

	// Is there a local by that name?
	topTypeScope := tc.currentScope
	ident, ok := topTypeScope.Locals[identName]
	if ok {
		tc.defUsages.Add(tc.makeCodeLocation(bounds), ident, true)
		tc.pushExprType(ident.Type, bounds)
		return
	}

	enclosing := tc.getEnclosingType()
	// Is there a class member by that name?
	member := enclosing.LookupMember(identName)
	if member != nil {
		tc.defUsages.Add(tc.makeCodeLocation(bounds), member, true)
		tc.pushExprType(member.GetType(), bounds)
		return
	}

	// Is there a type by that name?
	ttype := tc.lookupType(identName)
	if ttype != nil {
		tc.defUsages.Add(tc.makeCodeLocation(bounds), ttype, true)
		tc.pushExprType(ttype.GetType(), bounds)
		return
	}

	// Not found
	tc.addError(TypeError{
		Loc:     bounds,
		Message: fmt.Sprintf("Unknown identifier: %s", identName),
	})

	// The rest of the expression needs something to continue
	tc.pushAnyType(bounds)
}

func (tc *typeChecker) ExitCreator(ctx *javaparser.CreatorContext) {
	// TODO handle array creations eg. `var intArray = new int[5]`

	createdName := ctx.CreatedName().(*javaparser.CreatedNameContext)
	// TODO handle generics
	// TODO handle multiple identifiers, e.g. `new OuterClass.InnerClass()`
	identName := createdName.Identifier(0).GetText()

	createdType := tc.lookupType(identName)
	if createdType != nil {
		tc.defUsages.Add(tc.makeCodeLocation(loc.ParserRuleContextToBounds(ctx)), createdType, true)
	}

	// TODO constructor resolution & type checking

	tc.pushExprTypeName(identName, loc.ParserRuleContextToBounds(ctx))
}

func (tc *typeChecker) ExitMethodCall(ctx *javaparser.MethodCallContext) {
	if tc.insideExpressionType(ExprTypeDotExpr) {
		// If we're a method call inside of a dot expression (e.g. `System.exit()`), don't worry
		// about it, since it'll get handled by handleDotExpr()
		return
	}

	// TODO handle this()/super() method calls (allowed in constructors only I believe)
	ident := ctx.Identifier()
	if ident != nil {
		tc.handleIdentifier(ident.(*javaparser.IdentifierContext))
	}

	bounds := loc.ParserRuleContextToBounds(ctx)

	// The identifier should be a method of type __LSPMethod__
	methodType := tc.expressionStack.Pop().ttype
	if methodType.Type != typ.JavaTypeLSPMethod {
		//tc.logger.Error("method is not __LSPMethod__, instead it's: " + methodType.Name)
		tc.pushAnyType(bounds)
		return
	}

	tc.handleMethodCall(ctx, methodType, ident.GetText())
}

func (tc *typeChecker) handleMethodCall(ctx *javaparser.MethodCallContext, methodType *typ.JavaType, methodName string) {
	bounds := loc.ParserRuleContextToBounds(ctx)

	// At this point, all expressions should be in order on the expression stack,
	// from being previously visited.
	// We just need to pop them off one by one and make sure they are compatible with
	// the arguments of this method.
	paramTypes := methodType.GenericArgs[2:]
	// Iterate backwards since we'll pop them off the stack backwards
	foundArguments := 0
	for i := len(paramTypes) - 1; i >= 0; i-- {
		// param is the one in the function def
		paramType := paramTypes[i]

		// arg is the one in the function call
		argType := tc.expressionStack.Pop()

		if argType.ttype == nil {
			tc.addError(TypeError{
				Loc:     bounds,
				Message: fmt.Sprintf("Not enough arguments in function call to %s! Expected %d, got %d", methodName, len(paramTypes), foundArguments),
			})
		} else {
			foundArguments++

			if !argType.ttype.CoercesTo(paramType) {
				tc.addError(TypeError{
					Loc:     bounds,
					Message: fmt.Sprintf("Can't use %s as type %s in function call to %s", argType.ttype.ShortName(), paramType.ShortName(), methodName),
				})
			}
		}
	}

	// TODO check for too many arguments without clearing the entire expression stack

	// Now that the function call is resolved, push its return type onto the expression stack
	tc.pushExprType(methodType.GenericArgs[1], bounds)
}

func (tc *typeChecker) EnterExpression(ctx *javaparser.ExpressionContext) {
	dotToken := ctx.GetDotop()
	if dotToken != nil {
		tc.pushPlaceholder(ExprTypeDotExpr)
	}
}

func (tc *typeChecker) ExitExpression(ctx *javaparser.ExpressionContext) {
	dotToken := ctx.GetDotop()
	if dotToken != nil {
		tc.handleDotExpr(ctx)
	}

	bopToken := ctx.GetBop()
	if bopToken != nil {
		bop := bopToken.GetText()
		tc.handleBinaryExpression(bop, loc.ParserRuleContextToBounds(ctx))
	}
}

func (tc *typeChecker) handleDotExpr(ctx *javaparser.ExpressionContext) {
	// When entering the dot operator expression, we pushed a placeholder onto the stack.
	// Now we pop off everything until that placeholder so we can deal with it in a different order.
	exprs := tc.popUntilPlaceholderType(ExprTypeDotExpr)

	if len(exprs) == 0 {
		// TODO log this?
		tc.pushAnyType(loc.ParserRuleContextToBounds(ctx))
		return
	}

	left := exprs[len(exprs)-1]

	ident := ctx.Identifier()
	if ident != nil {
		// We're referring to a field (such as `System.in`)
		identName := ident.GetText()
		member := left.ttype.LookupMember(identName)

		var memberType *typ.JavaType

		if member == nil {
			tc.addError(TypeError{
				Loc:     loc.ParserRuleContextToBounds(ident),
				Message: fmt.Sprintf("Can't find member named %s of type %s", identName, left.ttype.ShortName()),
			})
			memberType = tc.lookupOrCreateType("any")
		} else {
			tc.defUsages.Add(tc.makeCodeLocation(loc.ParserRuleContextToBounds(ident)), member, true)
			memberType = member.GetType()
		}

		tc.pushExprType(memberType, loc.ParserRuleContextToBounds(ctx))
		return
	}

	methodCallI := ctx.MethodCall()
	if methodCallI != nil {
		// We're referring to a method (such as `System.exit()`).
		methodCall := methodCallI.(*javaparser.MethodCallContext)

		ident = methodCall.Identifier()
		if ident == nil {
			// TODO handle this/super
			return
		}

		identName := ident.GetText()
		member := left.ttype.LookupMember(identName)

		if member == nil {
			tc.addError(TypeError{
				Loc:     loc.ParserRuleContextToBounds(ident),
				Message: fmt.Sprintf("Can't find member named %s on type %s", identName, left.ttype.ShortName()),
			})
			tc.pushAnyType(loc.ParserRuleContextToBounds(ident))
			return
		}

		tc.defUsages.Add(tc.makeCodeLocation(loc.ParserRuleContextToBounds(ident)), member, true)

		methodType := member.GetType()
		if methodType.Type != typ.JavaTypeLSPMethod {
			bounds := loc.ParserRuleContextToBounds(methodCall)
			tc.addError(TypeError{
				Loc:     bounds,
				Message: fmt.Sprintf("%s is not callable", methodType.FullName()),
			})
			tc.pushAnyType(bounds)
			return
		}

		// The args were pushed onto the stack after `left`. handleMethodCall is going to typecheck
		// them, but it expects them back on the stack (in reverse order).
		args := util.Reverse(exprs[:len(exprs)-1])
		for _, arg := range args {
			tc.expressionStack.Push(arg)
		}
		tc.handleMethodCall(methodCall, methodType, left.ttype.GetClassName()+"."+identName)
	}

	// TODO handle other possibilities
}

// Binary operators that take in two numbers and return a number
var concatBop = util.SetFromValues("+", "+=")
var arithmeticBops = util.SetFromValues("+", "-", "*", "/", "%", "+=", "-=", "*=", "/=", "%=")
var bitshiftBops = util.SetFromValues(">>", ">>>", "<<", ">>=", ">>>=", "<<=")
var comparisonBops = util.SetFromValues("<", ">", "<=", ">=")
var bitwiseBops = util.SetFromValues("&", "|", "^", "&=", "|=", "^=")
var equalityBops = util.SetFromValues("==", "!=")
var booleanBops = util.SetFromValues("&&", "||", "&&=", "||=")

func (tc *typeChecker) handleBinaryExpression(bop string, exprBounds loc.Bounds) {
	right := tc.expressionStack.Pop()
	left := tc.expressionStack.Pop()

	exprNilFunc := func(side string) {
		tc.addError(TypeError{
			Message: fmt.Sprintf("TODO: %s expression is nil (this shouldn't happen, contact extension maintainers)", side),
			Loc:     exprBounds,
		})
		tc.expressionStack.Push(typedExpression{
			loc:                 exprBounds,
			ttype:               tc.lookupOrCreateType("any"),
			placeholderExprType: ExprTypeUnset,
		})
	}
	if right.ttype == nil {
		exprNilFunc("right")
		return
	}
	if left.ttype == nil {
		exprNilFunc("left")
		return
	}

	alwaysReturnString := func(_ *typ.JavaType, _ *typ.JavaType) *typ.JavaType {
		return tc.lookupOrCreateType("String")
	}
	alwaysReturnBoolean := func(_ *typ.JavaType, _ *typ.JavaType) *typ.JavaType {
		return tc.lookupOrCreateType("boolean")
	}

	opType := "unknown"
	assertionFunc := emptyTypeAssertion
	returnTypeFunc := alwaysReturnBoolean
	// Some of the operators can contain = but are not assignment operators
	definitelyNotAssignment := false

	// Actually do the bop checking
	if concatBop.Contains(bop) && (right.ttype.Name == "String" || left.ttype.Name == "String") {
		opType = "concatenation"
		assertionFunc = assertIsNumericOrString
		returnTypeFunc = alwaysReturnString
	} else if arithmeticBops.Contains(bop) {
		opType = "arithmetic"
		assertionFunc = assertIsNumeric
		returnTypeFunc = determineArithmeticBopReturnType
	} else if bitshiftBops.Contains(bop) {
		opType = "bitshift"
		assertionFunc = assertIsIntegral
		returnTypeFunc = determineArithmeticBopReturnType
	} else if bitwiseBops.Contains(bop) {
		opType = "bitwise"
		assertionFunc = assertIsIntegral
		returnTypeFunc = determineArithmeticBopReturnType
	} else if comparisonBops.Contains(bop) {
		definitelyNotAssignment = true
		opType = "comparison"
		assertionFunc = assertIsNumeric
		returnTypeFunc = alwaysReturnBoolean
	} else if equalityBops.Contains(bop) {
		definitelyNotAssignment = true
		opType = "equality"
		assertionFunc = emptyTypeAssertion
		returnTypeFunc = alwaysReturnBoolean
	} else if booleanBops.Contains(bop) {
		opType = "boolean"
		assertionFunc = assertIsBoolean
		returnTypeFunc = alwaysReturnBoolean
	}

	var returnType *typ.JavaType

	if !definitelyNotAssignment && strings.Contains(bop, "=") {
		// Assignment is sort of a special case.
		// Always returns the type of the left element
		returnType = left.ttype
	} else {
		returnType = tc.determineBopReturnType(left, right, opType, assertionFunc, returnTypeFunc)
	}

	tc.expressionStack.Push(typedExpression{
		loc:                 exprBounds,
		ttype:               returnType,
		placeholderExprType: ExprTypeUnset,
	})
}

func (tc *typeChecker) determineBopReturnType(left, right typedExpression, opType string, assertionFunc func(ttype *typ.JavaType) bool, returnTypeFunc func(left *typ.JavaType, right *typ.JavaType) *typ.JavaType) *typ.JavaType {
	// First check to make sure that both operands are valid types. If one is not, just return the other one.
	assertionErrorFunc := func(expr typedExpression) {
		tc.addError(TypeError{
			Message: fmt.Sprintf("Cannot use %s operator on %s", opType, expr.ttype.Name),
			Loc:     expr.loc,
		})
	}
	if !assertionFunc(right.ttype) {
		assertionErrorFunc(right)
		return left.ttype
	}
	if !assertionFunc(left.ttype) {
		assertionErrorFunc(left)
		return right.ttype
	}

	// Then invoke returnTypeFunc to figure out what the new return type should be.
	return returnTypeFunc(left.ttype, right.ttype)
}

func assertIsNumericOrString(ttype *typ.JavaType) bool {
	return ttype.Name == "String" || assertIsNumeric(ttype)
}

var numericTypes = util.SetFromValues("byte", "char", "short", "int", "long", "float", "double")

func assertIsNumeric(ttype *typ.JavaType) bool {
	return ttype.Type == typ.JavaTypePrimitive && numericTypes.Contains(ttype.Name)
}

var integralTypes = util.SetFromValues("byte", "char", "short", "int", "long")

func assertIsIntegral(ttype *typ.JavaType) bool {
	return ttype.Type == typ.JavaTypePrimitive && integralTypes.Contains(ttype.Name)
}

func assertIsBoolean(ttype *typ.JavaType) bool {
	return ttype.Type == typ.JavaTypePrimitive && ttype.Name == "boolean"
}

func emptyTypeAssertion(_ *typ.JavaType) bool {
	return true
}

// TODO char doesn't quite belong in this list, it's a bit of a special case
// I think if you add 2 chars they stay a char
// If you add a char to a short/byte it gets promoted to an int maybe? because char is unsigned
var integralTypeWidths = []string{"byte", "short", "char", "int", "long"}

// TODO this is for general arithmetic operators, it may be different for bitwise/bitshift ones
// (e.g. right now this accepts Strings)
func determineArithmeticBopReturnType(left *typ.JavaType, right *typ.JavaType) *typ.JavaType {
	// If either left or right is a double, the return value is a double
	if left.Name == "double" {
		return left
	}
	if right.Name == "double" {
		return right
	}

	// If left or right is a float, return value is a float
	if left.Name == "float" {
		return left
	}
	if right.Name == "float" {
		return right
	}

	// Finally, if both are integral types, we return whichever is wider
	return widerOfIntegralTypes(left, right)
}

func widerOfIntegralTypes(left *typ.JavaType, right *typ.JavaType) *typ.JavaType {
	idxOfLeft := slices.Index(integralTypeWidths, left.Name)
	idxOfRight := slices.Index(integralTypeWidths, right.Name)

	if idxOfLeft > idxOfRight {
		return left
	} else {
		return right
	}
}
