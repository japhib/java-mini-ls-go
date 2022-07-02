package parse

import (
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"go.lsp.dev/protocol"
	"golang.org/x/exp/slices"
	"java-mini-ls-go/javaparser"
	"java-mini-ls-go/util"
	"strings"
)

type TypeError struct {
	Loc     Bounds
	Message string
}

func (te *TypeError) ToDiagnostic() protocol.Diagnostic {
	return protocol.Diagnostic{
		Range:    BoundsToRange(te.Loc),
		Severity: protocol.DiagnosticSeverityError,
		Source:   "java-mini-ls",
		Message:  te.Message,
	}
}

// CheckTypes traverses the given parse tree and performs type checking in all applicable
// places. e.g. expressions, return statements, function calls, etc.
func CheckTypes(tree antlr.Tree, userTypes TypeMap, builtins TypeMap) []TypeError {
	visitor := &typeChecker{
		scopeTracker: NewScopeTracker(),
		userTypes:    userTypes,
		builtins:     builtins,
		errors:       make([]TypeError, 0),
		typeScopes:   util.NewStack[*typeCheckingScope](),
	}

	antlr.ParseTreeWalkerDefault.Walk(visitor, tree)

	return visitor.errors
}

type typedDeclarationCtx interface {
	TypeType() javaparser.ITypeTypeContext
	VariableDeclarators() javaparser.IVariableDeclaratorsContext
}

type localVar struct {
	name  string
	ttype *JavaType
}

type typeCheckingScope struct {
	locals map[string]localVar
}

func newTypeCheckingScope() *typeCheckingScope {
	return &typeCheckingScope{
		locals: make(map[string]localVar),
	}
}

func (tcs *typeCheckingScope) addLocal(name string, ttype *JavaType) {
	tcs.locals[name] = localVar{
		name:  name,
		ttype: ttype,
	}
}

type typedExpression struct {
	loc   Bounds
	ttype *JavaType
}

func (te typedExpression) String() string {
	return fmt.Sprintf("loc=%v type=%v", te.loc, te.ttype)
}

type typeChecker struct {
	javaparser.BaseJavaParserListener
	userTypes    TypeMap
	builtins     TypeMap
	errors       []TypeError
	scopeTracker *ScopeTracker
	typeScopes   util.Stack[*typeCheckingScope]

	// A stack used to keep track of the types of various expressions.
	// For example, in the binary expression `9 + 10`:
	// first, 9 will get evaluated, pushing the type "int" onto the stack
	// Then, 10 will get evaluated, pushing another "int" onto the stack
	// Then, the + binary expression gets evaluated, popping off the 2 last items
	// from the stack and checking their types.
	expressionStack util.Stack[typedExpression]
}

func (tc *typeChecker) addError(err TypeError) {
	tc.errors = append(tc.errors, err)
}

func (tc *typeChecker) lookupType(typeName string) *JavaType {
	userType, ok := tc.userTypes[typeName]
	if ok {
		return userType
	}

	if builtinType, ok := tc.builtins[typeName]; ok {
		return builtinType
	}

	// Type doesn't exist, create it
	// TODO checks getting from builtins again unnecessarily
	return getOrCreateBuiltinType(typeName)
}

func (tc *typeChecker) pushExprType(ttype *JavaType, bounds Bounds) {
	tc.expressionStack.Push(typedExpression{
		loc:   bounds,
		ttype: ttype,
	})
}

func (tc *typeChecker) pushExprTypeName(typeName string, bounds Bounds) {
	tc.pushExprType(tc.lookupType(typeName), bounds)
}

// checkAndAddLocal adds a local variable, while first checking whether the local
// is already defined, and if so, adding an error.
func (tc *typeChecker) checkAndAddLocal(name string, ttype *JavaType, bounds Bounds, scopeType string) {
	topScope := tc.typeScopes.Top()
	if _, ok := topScope.locals[name]; ok {
		currMethodName := tc.scopeTracker.ScopeStack.Top().Name
		tc.addError(TypeError{
			Loc:     bounds,
			Message: fmt.Sprintf("Variable %s is already defined in %s %s", name, scopeType, currMethodName),
		})
	}

	topScope.addLocal(name, ttype)
}

func (tc *typeChecker) getEnclosingType() *JavaType {
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
		typeScope := newTypeCheckingScope()

		if newScope.Type.IsMethodType() {
			// Add method params to locals.
			// First, look up method in types.
			enclosingType := tc.getEnclosingType()
			// TODO handle method overrides (same name)
			method := enclosingType.Methods[newScope.Name]

			for _, param := range method.Params {
				typeScope.addLocal(param.Name, param.Type)
			}
		}

		tc.typeScopes.Push(typeScope)
	}
}

func (tc *typeChecker) ExitEveryRule(ctx antlr.ParserRuleContext) {
	oldScope := tc.scopeTracker.CheckExitScope(ctx)
	if oldScope != nil {
		tc.typeScopes.Pop()
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
	tc.handleTypedVariableDecl(ctx, ParserRuleContextToBounds(ctx), false)
}

func (tc *typeChecker) ExitLocalVariableDeclaration(ctx *javaparser.LocalVariableDeclarationContext) {
	if ctx.VAR() == nil {
		tc.handleTypedVariableDecl(ctx, ParserRuleContextToBounds(ctx), true)
	} else {
		tc.handleUntypedLocalVariableDecl(ctx, ParserRuleContextToBounds(ctx))
	}
}

// e.g. `String a = "hi"`
func (tc *typeChecker) handleTypedVariableDecl(ctx typedDeclarationCtx, bounds Bounds, isLocal bool) {
	ttype := tc.lookupType(ctx.TypeType().GetText())

	// There can be multiple variable declarators
	varDecls := ctx.VariableDeclarators().(*javaparser.VariableDeclaratorsContext).AllVariableDeclarator()
	for _, varDeclI := range varDecls {
		varDecl := varDeclI.(*javaparser.VariableDeclaratorContext)

		varName := varDecl.VariableDeclaratorId().(*javaparser.VariableDeclaratorIdContext).Identifier().GetText()

		var scopeType string
		if isLocal {
			scopeType = "method"
		} else {
			scopeType = "class"
		}

		// TODO fix bounds, the error message also red underlines the equals sign
		tc.checkAndAddLocal(varName, ttype, bounds, scopeType)
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
func (tc *typeChecker) handleUntypedLocalVariableDecl(ctx *javaparser.LocalVariableDeclarationContext, bounds Bounds) {
	// In order for type to be inferred, we must have already pushed the expression type
	ttype := tc.expressionStack.Pop().ttype

	tc.checkAndAddLocal(ctx.Identifier().GetText(), ttype, bounds, "method")
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
	bounds := ParserRuleContextToBounds(ctx)

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
	bounds := ParserRuleContextToBounds(ctx)
	identName := ctx.GetText()

	// Is there a local by that name?
	topTypeScope := tc.typeScopes.Top()
	ident, ok := topTypeScope.locals[identName]
	if ok {
		tc.pushExprType(ident.ttype, bounds)
		return
	}

	enclosing := tc.getEnclosingType()
	field := enclosing.LookupField(identName)
	if field != nil {
		tc.pushExprType(field.Type, bounds)
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
		tc.handleBinaryExpression(bop, ParserRuleContextToBounds(ctx))
	}
}

// Binary operators that take in two numbers and return a number
var arithmeticBops = util.SetFromValues("+", "-", "*", "/", "%", "+=", "-=", "*=", "/=", "%=")
var bitshiftBops = util.SetFromValues(">>", ">>>", "<<", ">>=", ">>>=", "<<=")
var comparisonBops = util.SetFromValues("<", ">", "<=", ">=")
var bitwiseBops = util.SetFromValues("&", "|", "^", "&=", "|=", "^=")
var equalityBops = util.SetFromValues("==", "!=")
var booleanBops = util.SetFromValues("&&", "||", "&&=", "||=")

func (tc *typeChecker) handleBinaryExpression(bop string, exprBounds Bounds) {
	left := tc.expressionStack.Pop()
	right := tc.expressionStack.Pop()
	// TODO panic instead of adding error(s) when we're more confident
	if right.ttype == nil {
		tc.addError(TypeError{
			Message: "TODO: right expression is nil",
			Loc:     exprBounds,
		})
		tc.expressionStack.Push(typedExpression{
			loc:   exprBounds,
			ttype: tc.lookupType("Object"),
		})
		return
	}
	if left.ttype == nil {
		tc.addError(TypeError{
			Message: "TODO: left expression is nil",
			Loc:     exprBounds,
		})
		tc.expressionStack.Push(typedExpression{
			loc:   exprBounds,
			ttype: tc.lookupType("Object"),
		})
		return
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

	var returnType *JavaType

	if !definitelyNotAssignment && strings.Contains(bop, "=") {
		// Assignment is sort of a special case
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

func (tc *typeChecker) determineBopReturnType(left, right typedExpression, opType string, assertionFunc func(ttype *JavaType) bool, returnTypeFunc func(left *JavaType, right *JavaType) *JavaType) *JavaType {
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

func assertIsNumeric(ttype *JavaType) bool {
	return ttype.Type == JavaTypePrimitive && numericTypes.Contains(ttype.Name)
}

var integralTypes = util.SetFromValues("byte", "char", "short", "int", "long")

func assertIsIntegral(ttype *JavaType) bool {
	return ttype.Type == JavaTypePrimitive && integralTypes.Contains(ttype.Name)
}

func assertIsBoolean(ttype *JavaType) bool {
	return ttype.Type == JavaTypePrimitive && ttype.Name == "boolean"
}

func emptyTypeAssertion(_ *JavaType) bool {
	return true
}

// TODO char doesn't quite belong in this list, it's a bit of a special case
// I think if you add 2 chars they stay a char
// If you add a char to a short/byte it gets promoted to an int maybe? because char is unsigned
var integralTypeWidths = []string{"byte", "short", "char", "int", "long"}

// TODO this is for general arithmetic operators, it may be different for bitwise/bitshift ones
// (e.g. right now this accepts Strings)
func determineArithmeticBopReturnType(left *JavaType, right *JavaType) *JavaType {
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

func widerOfIntegralTypes(left *JavaType, right *JavaType) *JavaType {
	idxOfLeft := slices.Index(integralTypeWidths, left.Name)
	idxOfRight := slices.Index(integralTypeWidths, right.Name)

	if idxOfLeft > idxOfRight {
		return left
	} else {
		return right
	}
}

func alwaysReturnsBoolean(_ *JavaType, _ *JavaType) *JavaType {
	return getOrCreateBuiltinType("boolean")
}
