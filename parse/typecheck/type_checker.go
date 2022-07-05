package typecheck

import (
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"java-mini-ls-go/javaparser"
	"java-mini-ls-go/parse"
	"java-mini-ls-go/util"
	"math"
	"strings"
)

type TypeError struct {
	Loc     parse.Bounds
	Message string
}

func (te *TypeError) ToDiagnostic() protocol.Diagnostic {
	return protocol.Diagnostic{
		Range:    parse.BoundsToRange(te.Loc),
		Severity: protocol.DiagnosticSeverityError,
		Source:   "java-mini-ls",
		Message:  te.Message,
	}
}

//goland:noinspection GoNameStartsWithPackageName
type TypeCheckResult struct {
	TypeErrors      []TypeError
	DefUsagesLookup *DefinitionsUsagesLookup
	RootScope       TypeCheckingScope
}

// CheckTypes traverses the given parse tree and performs type checking in all applicable
// places. e.g. expressions, return statements, function calls, etc.
func CheckTypes(logger *zap.Logger, tree antlr.Tree, fileURI string, userTypes parse.TypeMap, builtins parse.TypeMap) TypeCheckResult {
	visitor := newTypeChecker(logger, fileURI, userTypes, builtins)
	antlr.ParseTreeWalkerDefault.Walk(visitor, tree)

	return TypeCheckResult{
		TypeErrors:      visitor.errors,
		DefUsagesLookup: visitor.defUsages,
		RootScope:       visitor.rootScope,
	}
}

type typedDeclarationCtx interface {
	TypeType() javaparser.ITypeTypeContext
	VariableDeclarators() javaparser.IVariableDeclaratorsContext
}

type typedExpression struct {
	loc   parse.Bounds
	ttype *parse.JavaType
}

func (te typedExpression) String() string {
	return fmt.Sprintf("loc=%v type=%v", te.loc, te.ttype)
}

type typeChecker struct {
	javaparser.BaseJavaParserListener
	logger       *zap.Logger
	currFileURI  string
	userTypes    parse.TypeMap
	builtins     parse.TypeMap
	errors       []TypeError
	scopeTracker *parse.ScopeTracker
	rootScope    TypeCheckingScope
	currentScope *TypeCheckingScope
	defUsages    *DefinitionsUsagesLookup

	// A stack used to keep track of the types of various expressions.
	// For example, in the binary expression `9 + 10`:
	// first, 9 will get evaluated, pushing the type "int" onto the stack
	// Then, 10 will get evaluated, pushing another "int" onto the stack
	// Then, the + binary expression gets evaluated, popping off the 2 last items
	// from the stack and checking their types.
	expressionStack util.Stack[typedExpression]
}

func newTypeChecker(logger *zap.Logger, fileURI string, userTypes parse.TypeMap, builtins parse.TypeMap) *typeChecker {
	rootScope := newTypeCheckingScope(
		nil,
		nil,
		// Bounds representing the entire file
		parse.Bounds{
			Start: parse.FileLocation{
				Line:   0,
				Column: 0,
			},
			End: parse.FileLocation{
				Line:   math.MaxInt,
				Column: math.MaxInt,
			},
		},
	)

	return &typeChecker{
		logger:          logger.Named("typeChecker"),
		currFileURI:     fileURI,
		userTypes:       userTypes,
		builtins:        builtins,
		errors:          make([]TypeError, 0),
		scopeTracker:    parse.NewScopeTracker(),
		rootScope:       rootScope,
		currentScope:    &rootScope,
		defUsages:       NewDefinitionsUsagesLookup(),
		expressionStack: util.NewStack[typedExpression](),
	}
}

func (tc *typeChecker) addError(err TypeError) {
	tc.errors = append(tc.errors, err)
}

func (tc *typeChecker) lookupType(typeName string) *parse.JavaType {
	userType, ok := tc.userTypes[typeName]
	if ok {
		return userType
	}

	if builtinType, ok := tc.builtins[typeName]; ok {
		return builtinType
	}

	// Type doesn't exist, create it
	fmt.Println("Creating built-in type: ", typeName)
	jtype := &parse.JavaType{
		Name:       typeName,
		Visibility: parse.VisibilityPublic,
	}
	tc.builtins[typeName] = jtype

	return jtype
}

func (tc *typeChecker) pushExprType(ttype *parse.JavaType, bounds parse.Bounds) {
	tc.expressionStack.Push(typedExpression{
		loc:   bounds,
		ttype: ttype,
	})
}

func (tc *typeChecker) pushExprTypeName(typeName string, bounds parse.Bounds) {
	tc.pushExprType(tc.lookupType(typeName), bounds)
}

// checkAndAddVariable adds a local variable, while first checking whether the local
// is already defined, and if so, adding an error.
func (tc *typeChecker) checkAndAddVariable(name string, ttype *parse.JavaType, bounds parse.Bounds, scopeType string) {
	topScope := tc.currentScope
	if _, ok := topScope.Locals[name]; ok {
		currMethodName := tc.scopeTracker.ScopeStack.Top().Name
		tc.addError(TypeError{
			Loc:     bounds,
			Message: fmt.Sprintf("Variable %s is already defined in %s %s", name, scopeType, currMethodName),
		})
	}

	enclosingMethod, ok := topScope.Symbol.(*parse.JavaMethod)
	if ok {
		// We're inside a method, so it's a local
		local := parse.NewJavaLocal(name, ttype, enclosingMethod, parse.CodeLocation{
			FileUri: tc.currFileURI,
			Loc:     bounds,
		})
		topScope.addLocal(local)
		tc.defUsages.NewSymbol(bounds, local)
	}
}

func (tc *typeChecker) getEnclosingType() *parse.JavaType {
	scopes := tc.scopeTracker.ScopeStack
	for i := scopes.Size() - 1; i >= 0; i-- {
		scope := scopes.TopMinus(i)
		if scope.Type.IsClassType() {
			return tc.lookupType(scope.Name)
		}
	}
	return nil
}

func (tc *typeChecker) EnterEveryRule(ctx antlr.ParserRuleContext) {
	newScope := tc.scopeTracker.CheckEnterScope(ctx)
	if newScope != nil {
		bounds := parse.ParserRuleContextToBounds(ctx)
		symbolForScope := tc.getSymbolFromScope(newScope)
		typeScope := newTypeCheckingScope(symbolForScope, tc.currentScope, bounds)

		if newScope.Type.IsMethodType() {
			// Add method params to Locals.
			// First, look up method in types.
			enclosingType := tc.getEnclosingType()

			// TODO handle method overrides (same name)
			methodIdx := slices.IndexFunc(enclosingType.Methods, func(method *parse.JavaMethod) bool {
				return method.Name == newScope.Name
			})
			method := enclosingType.Methods[methodIdx]

			for _, param := range method.Params {
				local := parse.NewJavaLocal(param.Name, param.Type, method, parse.CodeLocation{
					FileUri: tc.currFileURI,
					Loc:     bounds,
				})
				typeScope.addLocal(local)
			}
		}

		tc.currentScope = &typeScope
	}
}

func (tc *typeChecker) getSymbolFromScope(scope *parse.Scope) parse.JavaSymbol {
	if scope.Type.IsClassType() {
		return tc.lookupType(scope.Name)
	}

	// it's a method, should be inside a class
	if tc.currentScope.Symbol == nil {
		tc.logger.Error(fmt.Sprintf("can't get symbol from scope, tc.currentScope.Symbol is nil. Scope: %v", scope))
		return nil
	}

	enclosingType, ok := tc.currentScope.Symbol.(*parse.JavaType)
	if !ok {
		tc.logger.Error(fmt.Sprintf("can't get symbol from scope, enclosing type is %T. Scope: %v", tc.currentScope.Symbol, scope))
		return nil
	}

	// Look up method on enclosingType
	return enclosingType.LookupMethod(scope.Name)
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
	tc.handleTypedVariableDecl(ctx, parse.ParserRuleContextToBounds(ctx), false)
}

func (tc *typeChecker) ExitLocalVariableDeclaration(ctx *javaparser.LocalVariableDeclarationContext) {
	typedI := ctx.TypedLocalVarDecl()
	if typedI != nil {
		typed := typedI.(*javaparser.TypedLocalVarDeclContext)
		tc.handleTypedVariableDecl(typed, parse.ParserRuleContextToBounds(ctx), true)
	} else {
		untyped := ctx.UntypedLocalVarDecl().(*javaparser.UntypedLocalVarDeclContext)
		tc.handleUntypedLocalVariableDecl(untyped)
	}
}

// e.g. `String a = "hi"`
func (tc *typeChecker) handleTypedVariableDecl(ctx typedDeclarationCtx, bounds parse.Bounds, isLocal bool) {
	ttype := tc.lookupType(ctx.TypeType().GetText())

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
		tc.checkAndAddVariable(varName, ttype, parse.ParserRuleContextToBounds(ident), scopeType)
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

	tc.checkAndAddVariable(ctx.Identifier().GetText(), ttype, parse.ParserRuleContextToBounds(ctx.Identifier()), "method")
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
	bounds := parse.ParserRuleContextToBounds(ctx)

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
		// TODO add special "any" type that can be coerced to any type
		tc.pushExprTypeName("Object", bounds)
		return
	}

	textBlockLit := ctx.TEXT_BLOCK()
	if textBlockLit != nil {
		tc.pushExprTypeName("String", bounds)
		return
	}
}

func (tc *typeChecker) handleIdentifier(ctx *javaparser.IdentifierContext) {
	bounds := parse.ParserRuleContextToBounds(ctx)
	identName := ctx.GetText()

	// Is there a local by that name?
	topTypeScope := tc.currentScope
	ident, ok := topTypeScope.Locals[identName]
	if ok {
		tc.pushExprType(ident.Type, bounds)
		return
	}

	enclosing := tc.getEnclosingType()
	field := enclosing.LookupField(identName)
	if field != nil {
		tc.pushExprType(field.ParentType, bounds)
		return
	}

	// Not found
	tc.addError(TypeError{
		Loc:     bounds,
		Message: fmt.Sprintf("Unknown identifier: %s", identName),
	})

	// The rest of the expression needs something to continue -- we'll assume it's of type Object
	tc.pushExprType(tc.lookupType("Object"), bounds)
}

func (tc *typeChecker) ExitExpression(ctx *javaparser.ExpressionContext) {
	bopToken := ctx.GetBop()
	if bopToken != nil {
		bop := bopToken.GetText()
		tc.handleBinaryExpression(bop, parse.ParserRuleContextToBounds(ctx))
	}
}

// Binary operators that take in two numbers and return a number
var arithmeticBops = util.SetFromValues("+", "-", "*", "/", "%", "+=", "-=", "*=", "/=", "%=")
var bitshiftBops = util.SetFromValues(">>", ">>>", "<<", ">>=", ">>>=", "<<=")
var comparisonBops = util.SetFromValues("<", ">", "<=", ">=")
var bitwiseBops = util.SetFromValues("&", "|", "^", "&=", "|=", "^=")
var equalityBops = util.SetFromValues("==", "!=")
var booleanBops = util.SetFromValues("&&", "||", "&&=", "||=")

func (tc *typeChecker) handleBinaryExpression(bop string, exprBounds parse.Bounds) {
	right := tc.expressionStack.Pop()
	left := tc.expressionStack.Pop()

	exprNilFunc := func(side string) {
		tc.addError(TypeError{
			Message: fmt.Sprintf("TODO: %s expression is nil (this shouldn't happen, contact extension maintainers)", side),
			Loc:     exprBounds,
		})
		tc.expressionStack.Push(typedExpression{
			loc:   exprBounds,
			ttype: tc.lookupType("Object"),
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

	alwaysReturnsBoolean := func(_ *parse.JavaType, _ *parse.JavaType) *parse.JavaType {
		return tc.lookupType("boolean")
	}

	opType := "unknown"
	assertionFunc := emptyTypeAssertion
	returnTypeFunc := alwaysReturnsBoolean
	// Some of the operators can contain = but are not assignment operators
	definitelyNotAssignment := false

	// Actually do the bop checking
	if arithmeticBops.Contains(bop) {
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
		returnTypeFunc = alwaysReturnsBoolean
	} else if equalityBops.Contains(bop) {
		definitelyNotAssignment = true
		opType = "equality"
		assertionFunc = emptyTypeAssertion
		returnTypeFunc = alwaysReturnsBoolean
	} else if booleanBops.Contains(bop) {
		opType = "boolean"
		assertionFunc = assertIsBoolean
		returnTypeFunc = alwaysReturnsBoolean
	}

	var returnType *parse.JavaType

	if !definitelyNotAssignment && strings.Contains(bop, "=") {
		// Assignment is sort of a special case.
		// Always returns the type of the left element
		opType = "assignment"
		returnType = left.ttype
	} else {
		returnType = tc.determineBopReturnType(left, right, opType, assertionFunc, returnTypeFunc)
	}

	tc.expressionStack.Push(typedExpression{
		loc:   exprBounds,
		ttype: returnType,
	})
}

func (tc *typeChecker) determineBopReturnType(left, right typedExpression, opType string, assertionFunc func(ttype *parse.JavaType) bool, returnTypeFunc func(left *parse.JavaType, right *parse.JavaType) *parse.JavaType) *parse.JavaType {
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

var numericTypes = util.SetFromValues("byte", "char", "short", "int", "long", "float", "double")

func assertIsNumeric(ttype *parse.JavaType) bool {
	return ttype.Type == parse.JavaTypePrimitive && numericTypes.Contains(ttype.Name)
}

var integralTypes = util.SetFromValues("byte", "char", "short", "int", "long")

func assertIsIntegral(ttype *parse.JavaType) bool {
	return ttype.Type == parse.JavaTypePrimitive && integralTypes.Contains(ttype.Name)
}

func assertIsBoolean(ttype *parse.JavaType) bool {
	return ttype.Type == parse.JavaTypePrimitive && ttype.Name == "boolean"
}

func emptyTypeAssertion(_ *parse.JavaType) bool {
	return true
}

// TODO char doesn't quite belong in this list, it's a bit of a special case
// I think if you add 2 chars they stay a char
// If you add a char to a short/byte it gets promoted to an int maybe? because char is unsigned
var integralTypeWidths = []string{"byte", "short", "char", "int", "long"}

// TODO this is for general arithmetic operators, it may be different for bitwise/bitshift ones
// (e.g. right now this accepts Strings)
func determineArithmeticBopReturnType(left *parse.JavaType, right *parse.JavaType) *parse.JavaType {
	// If either left or right is a String, it's a string concatenation
	// so the return value is also a String
	if left.Name == "String" {
		return left
	}
	if right.Name == "String" {
		return right
	}

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

func widerOfIntegralTypes(left *parse.JavaType, right *parse.JavaType) *parse.JavaType {
	idxOfLeft := slices.Index(integralTypeWidths, left.Name)
	idxOfRight := slices.Index(integralTypeWidths, right.Name)

	if idxOfLeft > idxOfRight {
		return left
	} else {
		return right
	}
}
