package server

import (
	"context"
	"github.com/golang/mock/gomock"
	"java-mini-ls-go/util"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.uber.org/zap/zaptest"
)

func testCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func testServer(t *testing.T, ctx context.Context) (*JavaLS, *gomock.Controller) {
	jls := NewServer(ctx, zaptest.NewLogger(t))

	// Set up mocks
	ctrl := gomock.NewController(t)

	mdp := NewMockDiagnosticsPublisher(ctrl)
	jls.diagnosticsPublisher = mdp
	mdp.
		EXPECT().
		PublishDiagnostics(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes()

	fr := NewMockFileResolver(ctrl)
	jls.fileResolver = fr
	fr.
		EXPECT().
		FileURIToPath(gomock.Any()).
		DoAndReturn(func(fileUri string) (string, error) {
			return fileUri, nil
		}).
		AnyTimes()
	fr.
		EXPECT().
		ListJavaFilesRecursive(gomock.Any()).
		Return([]string{}, nil).
		AnyTimes()

	// Init server
	_, err := jls.Initialize(ctx, &protocol.InitializeParams{})
	assert.Nil(t, err)

	return jls, ctrl
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
	jls, _ := testServer(t, ctx)

	err := jls.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: createTextDocument("test_location", testFileText),
	})
	assert.Nil(t, err)
}

func TestServer_Symbols(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	jls, _ := testServer(t, ctx)

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

const shortTestFileText = `package java;

import somepkg.Thing;
import somepkg.nestedpkg.Nibble;`

func TestServer_getTextOnLine(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	jls, _ := testServer(t, ctx)

	err := jls.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: createTextDocument("test_location", shortTestFileText),
	})
	assert.Nil(t, err)

	// First line
	text, err := jls.getTextOnLine(string(uri.New("test_location")), 0)
	assert.Nil(t, err)
	assert.Equal(t, "package java;", text)

	// Some line in the middle
	text, err = jls.getTextOnLine(string(uri.New("test_location")), 2)
	assert.Nil(t, err)
	assert.Equal(t, "import somepkg.Thing;", text)

	// Last line
	text, err = jls.getTextOnLine(string(uri.New("test_location")), 3)
	assert.Nil(t, err)
	assert.Equal(t, "import somepkg.nestedpkg.Nibble;", text)
}

func TestServer_Hover(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	jls, _ := testServer(t, ctx)

	err := jls.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: createTextDocument("test_location", testFileText),
	})
	assert.Nil(t, err)

	// Hover in the middle of the "thing" variable declaration
	result, err := jls.Hover(ctx, &protocol.HoverParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri.New("test_location"),
			},
			Position: protocol.Position{
				Line:      13,
				Character: 14,
			},
		},
	})
	assert.Nil(t, err)

	assert.Equal(t, &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: "**thing** Thing thing (local var in void Main.main(String[] args))",
		},
		Range: nil,
	}, result)
}

const localTestFileText = `public class Main {
    public int main() {
		int a = 0;
		return a;
    }
}`

func TestServer_References(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	jls, _ := testServer(t, ctx)

	err := jls.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: createTextDocument("test_location", localTestFileText),
	})
	assert.Nil(t, err)

	result, err := jls.References(ctx, &protocol.ReferenceParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri.New("test_location"),
			},
			Position: protocol.Position{
				Line:      2,
				Character: 6,
			},
		},
	})
	assert.Nil(t, err)

	assert.Equal(t, []protocol.Location{
		{
			URI: uri.New("test_location"),
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      3,
					Character: 9,
				},
				End: protocol.Position{
					Line:      3,
					Character: 10,
				},
			},
		},
	}, result)
}

func TestServer_Definition(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	jls, _ := testServer(t, ctx)

	err := jls.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: createTextDocument("test_location", localTestFileText),
	})
	assert.Nil(t, err)

	result, err := jls.Definition(ctx, &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri.New("test_location"),
			},
			Position: protocol.Position{
				Line:      3,
				Character: 9,
			},
		},
	})
	assert.Nil(t, err)

	assert.Equal(t, []protocol.Location{
		{
			URI: uri.New("test_location"),
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      2,
					Character: 6,
				},
				End: protocol.Position{
					Line:      2,
					Character: 7,
				},
			},
		},
	}, result)
}

const localTestFileText2 = `public class Main {
	public void main() {
		int b = 1;
		long c = b;
		Integer boxedInt = b;
	}
}`

func TestServer_MultipleUsagesAndDef(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	jls, _ := testServer(t, ctx)

	err := jls.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: createTextDocument("test_location", localTestFileText2),
	})
	assert.Nil(t, err)

	// Get references -- should return 2
	result, err := jls.References(ctx, &protocol.ReferenceParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri.New("test_location"),
			},
			Position: protocol.Position{
				Line:      2,
				Character: 6,
			},
		},
	})
	assert.Nil(t, err)

	assert.Equal(t, []protocol.Location{
		{
			URI: uri.New("test_location"),
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      3,
					Character: 11,
				},
				End: protocol.Position{
					Line:      3,
					Character: 12,
				},
			},
		},
		{
			URI: uri.New("test_location"),
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      4,
					Character: 21,
				},
				End: protocol.Position{
					Line:      4,
					Character: 22,
				},
			},
		},
	}, result)

	// Get definition from 2 different places -- should both return the original declaration `int b = 1;`
	result, err = jls.Definition(ctx, &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri.New("test_location"),
			},
			Position: protocol.Position{
				Line:      3,
				Character: 11,
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, []protocol.Location{
		{
			URI: uri.New("test_location"),
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      2,
					Character: 6,
				},
				End: protocol.Position{
					Line:      2,
					Character: 7,
				},
			},
		},
	}, result)

	result, err = jls.Definition(ctx, &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri.New("test_location"),
			},
			Position: protocol.Position{
				Line:      4,
				Character: 21,
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, []protocol.Location{
		{
			URI: uri.New("test_location"),
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      2,
					Character: 6,
				},
				End: protocol.Position{
					Line:      2,
					Character: 7,
				},
			},
		},
	}, result)
}

func TestServer_Completion(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	jls, _ := testServer(t, ctx)

	err := jls.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: createTextDocument("test_location", `public class Main {
	public void main() {
		int Abcdef = 0;
		Abc
	}
}`)})
	assert.Nil(t, err)

	completionList, err := jls.Completion(ctx, &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri.New("test_location"),
			},
			Position: protocol.Position{
				Line:      3,
				Character: 5,
			},
		},
	})
	assert.Nil(t, err)

	assert.Equal(t, 2, len(completionList.Items))
	assert.Equal(t, "Abcdef", completionList.Items[0].Label)
	assert.Equal(t, "main", completionList.Items[1].Label)
}

func TestServer_Completion_Dot(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	jls, _ := testServer(t, ctx)

	err := jls.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: createTextDocument("test_location", `public class Main {
	public void main() {
		System.o
	}
}`)})
	assert.Nil(t, err)

	completionList, err := jls.Completion(ctx, &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri.New("test_location"),
			},
			Position: protocol.Position{
				Line:      2,
				Character: 12,
			},
		},
	})
	assert.Nil(t, err)

	// Should return "out" (as in System.out) as one of the results

	assert.NotEqual(t, 0, len(completionList.Items))
	foundOut := false
	for _, item := range completionList.Items {
		if item.Label == "out" {
			foundOut = true
		}
	}
	assert.True(t, foundOut)
}

func TestServer_Completion_DotIsLast(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	jls, _ := testServer(t, ctx)

	err := jls.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: createTextDocument("test_location", `public class Main {
	public void main() {
		System.out.
	}
}`)})
	assert.Nil(t, err)

	completionList, err := jls.Completion(ctx, &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri.New("test_location"),
			},
			Position: protocol.Position{
				Line:      2,
				Character: 13,
			},
		},
	})
	assert.Nil(t, err)

	// Should return "println" (as in System.out.println) as one of the results

	assert.NotEqual(t, 0, len(completionList.Items))
	foundOut := false
	for _, item := range completionList.Items {
		if item.Label == "println" {
			foundOut = true
		}
	}
	assert.True(t, foundOut)
}

func TestServer_ParseFilesOnStartup(t *testing.T) {
	ctx, cancel := testCtx()
	defer cancel()
	jls, ctrl := testServer(t, ctx)

	mdp := NewMockDiagnosticsPublisher(ctrl)
	jls.diagnosticsPublisher = mdp
	mdp.
		EXPECT().
		// Expect empty diagnostics list
		PublishDiagnostics(gomock.Any(), gomock.Any(), gomock.Eq([]protocol.Diagnostic{})).
		AnyTimes()

	mockClient := NewMockClient(ctrl)
	jls.client = mockClient
	mockClient.
		EXPECT().
		WorkspaceFolders(gomock.Any()).
		Return([]protocol.WorkspaceFolder{
			{
				URI:  "test_workspace_folder",
				Name: "test_workspace_folder",
			},
		}, nil).
		Times(1)

	fr := NewMockFileResolver(ctrl)
	jls.fileResolver = fr
	fr.
		EXPECT().
		FileURIToPath(gomock.Any()).
		DoAndReturn(func(fileUri string) (string, error) {
			return fileUri, nil
		}).
		AnyTimes()
	fr.
		EXPECT().
		ListJavaFilesRecursive(gomock.Eq("test_workspace_folder")).
		Return([]string{
			"abc.java",
			"def.java",
		}, nil).
		Times(1)
	fr.
		EXPECT().
		ReadFile(gomock.Eq("abc.java")).
		Return(`
public class Abc {
	public int a;

	public void DoAbc() {}

	public static void callDef() {
		var def = new Def();
		def.DoDef();
	}
}`).Times(1)

	fr.
		EXPECT().
		ReadFile(gomock.Eq("def.java")).
		Return(`
public class Def {
	public int d;

	public String DoDef() {
		return "def";
	}
}`).Times(1)

	err := jls.Initialized(ctx, nil)
	assert.Nil(t, err)

	err = jls.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
		TextDocument: createTextDocument("main.java", `public class Main {
	public void main() {
		var abc = new Abc();
		int b = abc.a;
		abc.DoAbc();
		
		var def = new Def();
		int e = def.d;
		String s = def.DoDef();
	}
}`)})
	assert.Nil(t, err)

	// Get definition of `abc.a`
	defResult, err := jls.Definition(ctx, &protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri.New("main.java"),
			},
			Position: protocol.Position{
				Line:      3,
				Character: 14,
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, []protocol.Location{
		{
			URI: uri.New("abc.java"),
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      2,
					Character: 12,
				},
				End: protocol.Position{
					Line:      2,
					Character: 13,
				},
			},
		},
	}, defResult)

	// Get usages of `def.DoDef()`
	refResult, err := jls.References(ctx, &protocol.ReferenceParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: uri.New("def.java"),
			},
			Position: protocol.Position{
				Line:      4,
				Character: 16,
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, []protocol.Location{
		{
			URI: uri.New("abc.java"),
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      8,
					Character: 6,
				},
				End: protocol.Position{
					Line:      8,
					Character: 11,
				},
			},
		},
		{
			URI: uri.New("main.java"),
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      8,
					Character: 17,
				},
				End: protocol.Position{
					Line:      8,
					Character: 22,
				},
			},
		},
	}, refResult)

}
