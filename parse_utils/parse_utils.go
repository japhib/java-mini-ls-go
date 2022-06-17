package parse_utils

import (
	"fmt"
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
		fmt.Printf("%s (%q)\n", lexer.SymbolicNames[t.GetTokenType()], t.GetText())
	}

	return tokens
}

type javaListener struct {
	*parser.BaseJavaParserListener
}

func (jl *javaListener) EnterClassDeclaration(ctx *parser.ClassDeclarationContext) {
	nameOfClass := ctx.Identifier().(antlr.ParseTree).GetText()
	fmt.Println("name of class ", nameOfClass)
}

func Parse(input string) {
	is := antlr.NewInputStream(input)
	lexer := parser.NewJavaLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	p := parser.NewJavaParser(stream)
	p.AddErrorListener(antlr.NewDiagnosticErrorListener(true))
	p.BuildParseTrees = true
	tree := p.CompilationUnit()
	//tree.ToStringTree()
	antlr.ParseTreeWalkerDefault.Walk(&javaListener{}, tree)
}
