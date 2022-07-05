package parse

import (
	"java-mini-ls-go/javaparser"
	"java-mini-ls-go/parse/loc"

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
		DefaultErrorListener: &antlr.DefaultErrorListener{},
		errors:               make([]SyntaxError, 0),
	}
	p.AddErrorListener(errListener)
	p.BuildParseTrees = true
	parsed := p.CompilationUnit().(*javaparser.CompilationUnitContext)

	return parsed, errListener.errors
}

type SyntaxError struct {
	Loc     loc.FileLocation
	Token   string
	Message string
}

func (se *SyntaxError) ToDiagnostic() protocol.Diagnostic {
	return protocol.Diagnostic{
		Range: loc.BoundsToRange(loc.Bounds{
			Start: se.Loc,
			End: loc.FileLocation{
				Line:   se.Loc.Line,
				Column: se.Loc.Column + len(se.Token),
			},
		}),
		Severity:           protocol.DiagnosticSeverityError,
		Code:               nil,
		CodeDescription:    nil,
		Source:             "java-mini-ls",
		Message:            se.Message,
		Tags:               []protocol.DiagnosticTag{},
		RelatedInformation: []protocol.DiagnosticRelatedInformation{},
		Data:               nil,
	}
}

type errorListener struct {
	*antlr.DefaultErrorListener
	errors []SyntaxError
}

var _ antlr.ErrorListener = (*errorListener)(nil)

func (el *errorListener) SyntaxError(_ antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, _ antlr.RecognitionException) {
	err := SyntaxError{
		Loc:     loc.FileLocation{Line: line, Column: column},
		Token:   offendingSymbol.(antlr.Token).GetText(),
		Message: msg,
	}
	el.errors = append(el.errors, err)
}
