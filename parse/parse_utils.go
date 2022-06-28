package parse

import (
	"java-mini-ls-go/javaparser"

	"go.lsp.dev/protocol"

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

func Parse(input string) (*javaparser.CompilationUnitContext, []SyntaxError) {
	is := antlr.NewInputStream(input)
	lexer := javaparser.NewJavaLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	p := javaparser.NewJavaParser(stream)
	errListener := &errorListener{
		errors: make([]SyntaxError, 0),
	}
	p.AddErrorListener(errListener)
	p.BuildParseTrees = true
	parsed := p.CompilationUnit().(*javaparser.CompilationUnitContext)

	return parsed, errListener.errors
}

type SyntaxError struct {
	Loc     FileLocation
	Token   string
	Message string
}

func (se *SyntaxError) ToDiagnostic() protocol.Diagnostic {
	return protocol.Diagnostic{
		Range: BoundsToRange(Bounds{
			Start: se.Loc,
			End:   FileLocation{se.Loc.Line, se.Loc.Column + len(se.Token)},
		}),
		Severity: protocol.DiagnosticSeverityError,
		Source:   "java-mini-ls",
		Message:  se.Message,
	}
}

type errorListener struct {
	*antlr.DefaultErrorListener
	errors []SyntaxError
}

var _ antlr.ErrorListener = (*errorListener)(nil)

func (el *errorListener) SyntaxError(_ antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, _ antlr.RecognitionException) {
	err := SyntaxError{
		Loc:     FileLocation{line, column},
		Token:   offendingSymbol.(antlr.Token).GetText(),
		Message: msg,
	}
	el.errors = append(el.errors, err)
}
