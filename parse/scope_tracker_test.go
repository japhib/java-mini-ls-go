package parse

import (
	"java-mini-ls-go/javaparser"
	"testing"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/stretchr/testify/assert"
)

type scope1 struct {
	name string
}

type scopeCreator1 struct{}

var _ ScopeCreator[scope1] = (*scopeCreator1)(nil)

func (sc *scopeCreator1) ShouldCreateScope(ruleType int) bool {
	switch ruleType {
	case javaparser.JavaParserRULE_classDeclaration:
		return true
	}
	return false
}

func (sc *scopeCreator1) CreateScope(ctx antlr.ParserRuleContext) *scope1 {
	ret := &scope1{}

	switch ctx.GetRuleIndex() {
	case javaparser.JavaParserRULE_classDeclaration:
		ret.name = ctx.(*javaparser.ClassDeclarationContext).Identifier().GetText()
	}

	return ret
}

type scopedVisitor1 struct {
	javaparser.BaseJavaParserListener
	scopeTracker *ScopeTracker[scope1]
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
	assert.Equal(s.t, "MyClass", scopes[0].name)
	assert.Equal(s.t, "Nested", scopes[1].name)
}

func TestScopedVisitor(t *testing.T) {
	tree := Parse(`
class MyClass{ 
	class Nested{ 
		public String asdf; 
	} 
}`)

	visitor := scopedVisitor1{
		scopeTracker: NewScopeTracker[scope1](&scopeCreator1{}),
		t:            t,
	}
	antlr.ParseTreeWalkerDefault.Walk(&visitor, tree)

	assert.Equal(t, 0, len(visitor.scopeTracker.ScopeStack))
}
