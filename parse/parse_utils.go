package parse

import (
	"java-mini-ls-go/javaparser"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

func Lex(input string) []antlr.Token {
	is := antlr.NewInputStream(input)
	lexer := javaparser.NewJavaLexer(is)

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

func Parse(input string) *javaparser.CompilationUnitContext {
	is := antlr.NewInputStream(input)
	lexer := javaparser.NewJavaLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	p := javaparser.NewJavaParser(stream)
	p.AddErrorListener(antlr.NewDiagnosticErrorListener(true))
	p.BuildParseTrees = true
	return p.CompilationUnit().(*javaparser.CompilationUnitContext)
}

func GetClassDeclBounds(*javaparser.ClassDeclarationContext) Bounds {
	return Bounds{}
}
