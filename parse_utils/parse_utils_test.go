package parse_utils

import (
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/stretchr/testify/assert"
	"java-mini-ls-go/parser"
	"testing"
)

type listener1 struct {
	*parser.BaseJavaParserListener
	nameOfClass string
}

func (l *listener1) EnterClassDeclaration(ctx *parser.ClassDeclarationContext) {
	l.nameOfClass = ctx.Identifier().(antlr.ParseTree).GetText()
}

func TestBasicParsing(t *testing.T) {
	parsed := Parse("class MyClass{}")
	listener := &listener1{}
	antlr.ParseTreeWalkerDefault.Walk(listener, parsed)
	assert.Equal(t, "MyClass", listener.nameOfClass)
}

type expectedToken struct {
	ttype int
	ttext string
}

func TestLex(t *testing.T) {
	tokens := Lex("class MyClass{}")

	expected := []expectedToken{
		{parser.JavaLexerCLASS, "class"},
		{parser.JavaLexerWS, " "},
		{parser.JavaLexerIDENTIFIER, "MyClass"},
		{parser.JavaLexerLBRACE, "{"},
		{parser.JavaLexerRBRACE, "}"},
		{antlr.TokenEOF, "<EOF>"},
	}

	for i := range tokens {
		assert.Equalf(t, expected[i].ttype, tokens[i].GetTokenType(), "token type not equal for index %d", i)
		assert.Equalf(t, expected[i].ttext, tokens[i].GetText(), "token text not equal for index %d", i)
	}
}
