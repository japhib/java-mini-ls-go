package parse

import (
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"go.lsp.dev/protocol"
)

type FileLocation struct {
	Line   int
	Column int
}

func FileLocationFromToken(token antlr.Token) FileLocation {
	return FileLocation{
		Line:   token.GetLine(),
		Column: token.GetColumn(),
	}
}

func (fl FileLocation) Equals(other FileLocation) bool {
	return fl.Line == other.Line &&
		fl.Column == other.Column
}

func (fl FileLocation) String() string {
	return fmt.Sprintf("%d:%d", fl.Line, fl.Column)
}

type CodeLocation struct {
	Loc      FileLocation
	Filename string
}

func (cl CodeLocation) Equals(other CodeLocation) bool {
	return cl.Filename == other.Filename &&
		cl.Loc.Equals(other.Loc)
}

func (cl CodeLocation) String() string {
	return fmt.Sprintf("%s#%s", cl.Filename, cl.Loc.String())
}

type Bounds struct {
	Start    FileLocation
	End      FileLocation
	Filename string
}

func BoundsToRange(bounds Bounds) protocol.Range {
	return protocol.Range{
		Start: protocol.Position{
			// Subtract 1 since we use 1-based line numbers but LSP expects 0-based
			Line:      (uint32)(bounds.Start.Line) - 1,
			Character: (uint32)(bounds.Start.Column),
		},
		End: protocol.Position{
			Line:      (uint32)(bounds.End.Line) - 1,
			Character: (uint32)(bounds.End.Column),
		},
	}
}

func (b Bounds) Equals(other Bounds) bool {
	return b.Filename == other.Filename &&
		b.Start.Equals(other.Start) &&
		b.End.Equals(other.End)
}

func (b Bounds) String() string {
	var loc string
	if b.Start.Line != b.End.Line {
		loc = fmt.Sprintf("%d:%d-%d:%d", b.Start.Line, b.Start.Column, b.End.Line, b.End.Column)
	} else {
		loc = fmt.Sprintf("%d:%d-%d", b.Start.Line, b.Start.Column, b.End.Column)
	}

	return fmt.Sprintf("%s#%s", b.Filename, loc)
}
