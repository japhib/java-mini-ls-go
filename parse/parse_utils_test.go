package parse

import (
	"java-mini-ls-go/javaparser"
	"testing"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/stretchr/testify/assert"
)

type listener1 struct {
	*javaparser.BaseJavaParserListener
	nameOfClass string
}

func (l *listener1) EnterClassDeclaration(ctx *javaparser.ClassDeclarationContext) {
	l.nameOfClass = ctx.Identifier().GetText()
}

func TestParse_Basic(t *testing.T) {
	parsed, errors := Parse("class MyClass{}")
	listener := &listener1{}
	antlr.ParseTreeWalkerDefault.Walk(listener, parsed)
	assert.Equal(t, 0, len(errors))
	assert.Equal(t, "MyClass", listener.nameOfClass)
}

func TestParse_SyntaxError(t *testing.T) {
	_, errors := Parse("class MyClass")
	assert.Equal(t, 1, len(errors))
}

type expectedToken struct {
	ttype int
	ttext string
}

func TestLex(t *testing.T) {
	tokens := Lex("class MyClass{}")

	expected := []expectedToken{
		{javaparser.JavaLexerCLASS, "class"},
		{javaparser.JavaLexerWS, " "},
		{javaparser.JavaLexerIDENTIFIER, "MyClass"},
		{javaparser.JavaLexerLBRACE, "{"},
		{javaparser.JavaLexerRBRACE, "}"},
		{antlr.TokenEOF, "<EOF>"},
	}

	for i := range tokens {
		assert.Equalf(t, expected[i].ttype, tokens[i].GetTokenType(), "token type not equal for index %d", i)
		assert.Equalf(t, expected[i].ttext, tokens[i].GetText(), "token text not equal for index %d", i)
	}
}
