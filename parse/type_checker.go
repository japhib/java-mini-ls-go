package parse

import (
	"fmt"
	"java-mini-ls-go/javaparser"
	"java-mini-ls-go/util"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"golang.org/x/exp/slices"
)

type TypeError struct {
	Loc     Bounds
	Message string
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

	return tc.builtins[typeName]
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
	tc.handleTypedVariableDecl(ctx)
}

func (tc *typeChecker) ExitLocalVariableDeclaration(ctx *javaparser.LocalVariableDeclarationContext) {
	if ctx.VAR() == nil {
		tc.handleTypedVariableDecl(ctx)
	} else {
		tc.handleUntypedLocalVariableDecl(ctx)
	}
}

// e.g. `String a = "hi"`
func (tc *typeChecker) handleTypedVariableDecl(ctx typedDeclarationCtx) {
	ttype := tc.lookupType(ctx.TypeType().GetText())
	currTypeScope := tc.typeScopes.Top()

	// There can be multiple variable declarators
	varDecls := ctx.VariableDeclarators().(*javaparser.VariableDeclaratorsContext).AllVariableDeclarator()
	for _, varDeclI := range varDecls {
		varDecl := varDeclI.(*javaparser.VariableDeclaratorContext)

		varName := varDecl.VariableDeclaratorId().(*javaparser.VariableDeclaratorIdContext).Identifier().GetText()
		currTypeScope.addLocal(varName, ttype)
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
func (tc *typeChecker) handleUntypedLocalVariableDecl(ctx *javaparser.LocalVariableDeclarationContext) {
	// In order for type to be inferred, we must have already pushed the expression type
	ttype := tc.expressionStack.Pop().ttype

	currTypeScope := tc.typeScopes.Top()
	currTypeScope.addLocal(ctx.Identifier().GetText(), ttype)
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
		tc.pushExprTypeName("int", bounds)
		return
	}

	floatLit := ctx.FloatLiteral()
	if floatLit != nil {
		tc.pushExprTypeName("float", bounds)
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
		tc.handleBinaryExpression(bop)
	}
}

// Binary operators that take in two numbers and return a number
var arithmeticBops = []string{"+", "-", "*", "/", "%"}

func (tc *typeChecker) handleBinaryExpression(bop string) {
	right := tc.expressionStack.Pop()
	if right.ttype == nil {
		panic("right ttype is nil")
	}
	left := tc.expressionStack.Pop()
	if left.ttype == nil {
		panic("left ttype is nil")
	}

	if slices.Contains(arithmeticBops, bop) {
		// right and left must both be numeric, and they will return a number
		tc.assertIsNumeric(left)
		tc.assertIsNumeric(right)
	}
}

var numericTypes = []string{"byte", "char", "short", "int", "long", "float", "double"}

func (tc *typeChecker) assertIsNumeric(expression typedExpression) {
	ttype := expression.ttype
	isNumeric := ttype.Type == JavaTypePrimitive && slices.Contains(numericTypes, ttype.Name)

	if !isNumeric {
		tc.addError(TypeError{
			Message: "Expected numeric type, instead got expression of type " + ttype.Name,
			Loc:     expression.loc,
		})
	}
}
