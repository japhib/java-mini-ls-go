package parse

import (
	"java-mini-ls-go/javaparser"
	"testing"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/stretchr/testify/assert"
)

type scopedVisitor1 struct {
	javaparser.BaseJavaParserListener
	scopeTracker *ScopeTracker
	t            *testing.T
}

var _ javaparser.JavaParserListener = (*scopedVisitor1)(nil)

func (s *scopedVisitor1) EnterEveryRule(ctx antlr.ParserRuleContext) {
	s.scopeTracker.CheckEnterScope(ctx)
}

func (s *scopedVisitor1) ExitEveryRule(ctx antlr.ParserRuleContext) {
	s.scopeTracker.CheckExitScope(ctx)
}

func (s *scopedVisitor1) EnterFieldDeclaration(ctx *javaparser.FieldDeclarationContext) {
	decls := ctx.VariableDeclarators().(*javaparser.VariableDeclaratorsContext).AllVariableDeclarator()
	assert.Equal(s.t, 1, len(decls))

	decl := decls[0]
	varName := decl.(*javaparser.VariableDeclaratorContext).VariableDeclaratorId().(*javaparser.VariableDeclaratorIdContext).Identifier().GetText()
	assert.Equal(s.t, "asdf", varName)

	scopes := s.scopeTracker.ScopeStack
	assert.Equal(s.t, "MyClass", scopes.TopMinus(1).Name)
	assert.Equal(s.t, "Nested", scopes.Top().Name)
}

func TestScopedVisitor(t *testing.T) {
	tree, errors := Parse(`
class MyClass{ 
	class Nested{ 
		public String asdf; 
	} 
}`)
	assert.Equal(t, 0, len(errors))

	visitor := scopedVisitor1{
		scopeTracker: NewScopeTracker(),
		t:            t,
	}
	antlr.ParseTreeWalkerDefault.Walk(&visitor, tree)

	assert.Equal(t, 0, visitor.scopeTracker.ScopeStack.Size())
}
