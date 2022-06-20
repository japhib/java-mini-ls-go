package parse_utils

import (
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"java-mini-ls-go/parser"
)

func Lex(input string) []antlr.Token {
	is := antlr.NewInputStream(input)
	lexer := parser.NewJavaLexer(is)

	// read all tokens
	tokens := make([]antlr.Token, 0)
	for {
		t := lexer.NextToken()
		tokens = append(tokens, t)

		if t.GetTokenType() == antlr.TokenEOF {
			break
		}
	}

	return tokens
}

func Parse(input string) *parser.CompilationUnitContext {
	is := antlr.NewInputStream(input)
	lexer := parser.NewJavaLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	p := parser.NewJavaParser(stream)
	p.AddErrorListener(antlr.NewDiagnosticErrorListener(true))
	p.BuildParseTrees = true
	return p.CompilationUnit().(*parser.CompilationUnitContext)
}
