package typecheck

import (
	"java-mini-ls-go/parse/loc"
	"java-mini-ls-go/parse/typ"
)

// SymbolWithLocation is a single definition or usage of a symbol.
// Contains the location of the definition/usage, and a pointer to the symbol
// which can be used to find other definitions/usages.
//
// Note that it uses Bounds instead of CodeLocation since it only is for a single file.
type SymbolWithLocation struct {
	Loc    loc.Bounds
	Symbol typ.JavaSymbol
}

// SymbolsOnLine is a list of all the identifiers on a line of code in the editor,
// their bounds, and all the SymbolWithLocation they point to.
type SymbolsOnLine []SymbolWithLocation

// DefinitionsUsagesLookup is a lookup table that helps the language server go from just
// a file URI and code location to figuring out what identifier is being pointed to,
// and finding its corresponding SymbolWithDefUsages struct.
type DefinitionsUsagesLookup struct {
	// DefUsagesByLine is a map of line numbers to the list of DefinitionsUsagesWithLocation on that line.
	DefUsagesByLine map[int]SymbolsOnLine
}

func NewDefinitionsUsagesLookup() *DefinitionsUsagesLookup {
	return &DefinitionsUsagesLookup{
		DefUsagesByLine: make(map[int]SymbolsOnLine),
	}
}

func (dul *DefinitionsUsagesLookup) GetLine(line int) SymbolsOnLine {
	return dul.DefUsagesByLine[line]
}

func (dul *DefinitionsUsagesLookup) NewSymbol(loc loc.Bounds, symbolToAdd typ.JavaSymbol) {
	lineNumber := loc.Start.Line
	line := dul.DefUsagesByLine[lineNumber]

	// If a list for that line doesn't exist, create it now
	if line != nil {
		line = make(SymbolsOnLine, 0)
		dul.DefUsagesByLine[lineNumber] = line
	}

	// Also make sure there's not already an item with that name/bounds
	for _, defUsages := range line {
		if defUsages.Symbol.GetDefinition() == symbolToAdd.GetDefinition() && defUsages.Loc.Equals(loc) {
			// Name/location match
			// TODO merge usages
			return
		}
	}

	// If we've made it this far, it hasn't been found so we can add it
	line = append(line, SymbolWithLocation{
		Symbol: symbolToAdd,
		Loc:    loc,
	})

	// Make sure to assign it back in case append had to realloc
	dul.DefUsagesByLine[lineNumber] = line
}

// Lookup Given a file location, returns the most specific SymbolWithDefUsages instance corresponding
// to that file location, if one exists.
func (dul *DefinitionsUsagesLookup) Lookup(loc loc.FileLocation) typ.JavaSymbol {
	line := dul.DefUsagesByLine[loc.Line]
	if line == nil {
		return nil
	}

	var found *SymbolWithLocation = nil

	for _, def := range line {
		matches := loc.Line == def.Loc.Start.Line &&
			loc.Column >= def.Loc.Start.Column &&
			loc.Column <= def.Loc.End.Column

		// Make sure we find the match with the narrowest bounds
		narrower := found == nil || def.Loc.Size() < found.Loc.Size()

		if matches && narrower {
			found = &def
		}
	}

	if found != nil {
		return found.Symbol
	}
	return nil
}
