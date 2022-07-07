package loc

import (
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"go.lsp.dev/protocol"
)

type FileLocation struct {
	Line      int
	Character int
}

func (fl FileLocation) Equals(other FileLocation) bool {
	return fl.Line == other.Line &&
		fl.Character == other.Character
}

func (fl FileLocation) String() string {
	return fmt.Sprintf("%d:%d", fl.Line, fl.Character)
}

type Bounds struct {
	Start FileLocation
	End   FileLocation
}

func ParserRuleContextToBounds(ctx antlr.ParserRuleContext) Bounds {
	startToken := ctx.GetStart()
	stopToken := ctx.GetStop()

	return Bounds{
		Start: FileLocation{startToken.GetLine(), startToken.GetColumn()},
		End:   FileLocation{stopToken.GetLine(), stopToken.GetColumn() + len(stopToken.GetText())},
	}
}

func BoundsToRange(bounds Bounds) protocol.Range {
	return protocol.Range{
		Start: protocol.Position{
			// Subtract 1 since we use 1-based line numbers but LSP expects 0-based
			Line:      uint32(bounds.Start.Line) - 1,
			Character: uint32(bounds.Start.Character),
		},
		End: protocol.Position{
			Line:      uint32(bounds.End.Line) - 1,
			Character: uint32(bounds.End.Character),
		},
	}
}

func (b Bounds) Equals(other Bounds) bool {
	return b.Start.Equals(other.Start) && b.End.Equals(other.End)
}

func (b Bounds) String() string {
	if b.Start.Line != b.End.Line {
		return fmt.Sprintf("%d:%d-%d:%d", b.Start.Line, b.Start.Character, b.End.Line, b.End.Character)
	} else {
		return fmt.Sprintf("%d:%d-%d", b.Start.Line, b.Start.Character, b.End.Character)
	}
}

func (b Bounds) Size() int {
	lineSize := b.End.Line - b.Start.Line

	columnSize := 0
	if lineSize == 0 {
		columnSize = b.End.Character - b.Start.Character
	}

	return (lineSize * 10000) + columnSize
}

type CodeLocation struct {
	FileUri string
	// Version is the version of this file -- used for checking whether the file contents are out-of-date
	Version int
	Loc     Bounds
}

func (cl CodeLocation) String() string {
	return fmt.Sprintf("%s|%d|%s", cl.FileUri, cl.Version, cl.Loc.String())
}

func (cl CodeLocation) Equals(other CodeLocation) bool {
	return cl.FileUri == other.FileUri && cl.Loc.Equals(other.Loc)
}
