package server

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.uber.org/zap/zaptest"
	"java-mini-ls-go/util"
	"testing"
	"time"
)

func testCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func testServer(t *testing.T, ctx context.Context) *JavaLS {
	jls := NewServer(ctx, zaptest.NewLogger(t))

	_, err := jls.Initialize(ctx, &protocol.InitializeParams{})
	assert.Nil(t, err)

	return jls
}

func createTextDocument(uriStr string, contents string) protocol.TextDocumentItem {
	return protocol.TextDocumentItem{
		URI:        uri.New(uriStr),
		Text:       contents,
		LanguageID: "java",
		Version:    0,
	}
}

const testFileText = `
package java;

import somepkg.Thing;
import somepkg.nestedpkg.Nibble;

// declares a type
public class Main {
    // declares a method on type Main
    public static void main(String[] args) {
        // function call "println" on type of "System.out" using string arg
        System.out.println("Hi there");

        Thing thing = new Thing("pub!", 3);
        System.out.println("the value of pubfield is " + thing.pubfield);
        System.out.println("the value of privfield is " + thing.getPrivField());

        Nibble nibble = new Nibble("asdf", 5);
        System.out.println("nibble: " + nibble);
    }
}`

func makeRange(startLine uint32, startCol uint32, endLine uint32, endCol uint32) protocol.Range {
	return protocol.Range{
		Start: protocol.Position{
			Line:      startLine,
			Character: startCol,
		},
		End: protocol.Position{
			Line:      endLine,
			Character: endCol,
		},
	}
}

func oneLineRange(line uint32, startCol uint32, endCol uint32) protocol.Range {
	return protocol.Range{
		Start: protocol.Position{
			Line:      line,
			Character: startCol,
		},
		End: protocol.Position{
			Line:      line,
			Character: endCol,
		},
	}
}

func TestServer_DidOpen(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	jls := testServer(t, ctx)

	err := jls.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: createTextDocument("test_location", testFileText),
	})
	assert.Nil(t, err)
}

func TestServer_Symbols(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	jls := testServer(t, ctx)

	err := jls.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: createTextDocument("test_location", testFileText),
	})
	assert.Nil(t, err)

	symbols, err := jls.DocumentSymbol(ctx, &protocol.DocumentSymbolParams{
		TextDocument: protocol.TextDocumentIdentifier{uri.New("test_location")},
	})
	assert.Nil(t, err)

	symbolsConverted := util.Map(symbols, func(i interface{}) protocol.DocumentSymbol {
		assert.NotNil(t, i)
		return i.(protocol.DocumentSymbol)
	})

	// TODO the ranges on the resolved symbols are a bit wonky, needs fixing
	expected := []protocol.DocumentSymbol{
		{
			Name:           "java",
			Kind:           protocol.SymbolKindPackage,
			Range:          oneLineRange(1, 0, 12),
			SelectionRange: oneLineRange(1, 0, 12),
			Children:       []protocol.DocumentSymbol{},
		},
		{
			Name:           "Main",
			Kind:           protocol.SymbolKindClass,
			Range:          makeRange(7, 7, 20, 0),
			SelectionRange: makeRange(7, 7, 20, 0),
			Children: []protocol.DocumentSymbol{
				{
					Name:           "main",
					Kind:           protocol.SymbolKindMethod,
					Range:          makeRange(9, 18, 19, 4),
					SelectionRange: makeRange(9, 18, 19, 4),
					Children: []protocol.DocumentSymbol{
						{
							Name: "args",
							Kind: protocol.SymbolKindVariable,
							// zero-character range on all these vars??
							Range:          makeRange(9, 37, 9, 37),
							SelectionRange: makeRange(9, 37, 9, 37),
							Children:       []protocol.DocumentSymbol{},
						},
						{
							Name:           "thing",
							Kind:           protocol.SymbolKindVariable,
							Range:          makeRange(13, 14, 13, 14),
							SelectionRange: makeRange(13, 14, 13, 14),
							Children:       []protocol.DocumentSymbol{},
						},
						{
							Name:           "nibble",
							Kind:           protocol.SymbolKindVariable,
							Range:          makeRange(17, 15, 17, 15),
							SelectionRange: makeRange(17, 15, 17, 15),
							Children:       []protocol.DocumentSymbol{},
						},
					},
				},
			},
		},
	}
	assert.Equal(t, expected, symbolsConverted)
}
