package main

import (
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"java-mini-ls-go/parser"
	"testing"
)

func TestParse(t *testing.T) {
	is := antlr.NewInputStream("class MyClass {}")

	lexer := parser.NewJavaLexer(is)

	// read all tokens
	for {
		t := lexer.NextToken()
		if t.GetTokenType() == antlr.TokenEOF {
			break
		}
		fmt.Printf("%s (%q)\n", lexer.SymbolicNames[t.GetTokenType()], t.GetText())
	}
}
