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
//
// Note: if there is a super long line (minified or something), this will probably have bad performance.
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

func (dul *DefinitionsUsagesLookup) Add(location loc.CodeLocation, symbol typ.JavaSymbol, addUsage bool) {
	if addUsage {
		symbol.AddUsage(location)
	}

	lineNumber := location.Loc.Start.Line
	// Look up the list for the line
	// Note: it's okay if it's nil because append treats it the same as if it were an empty slice.
	line := dul.DefUsagesByLine[lineNumber]

	// Also make sure there's not already an item with that name/bounds
	for _, defUsages := range line {
		if defUsages.Symbol.GetDefinition() == symbol.GetDefinition() && defUsages.Loc.Equals(location.Loc) {
			// Name/location match
			// TODO merge usages
			return
		}
	}

	// If we've made it this far, it hasn't been found so we can add it
	line = append(line, SymbolWithLocation{
		Symbol: symbol,
		Loc:    location.Loc,
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

	alreadyFound := false
	var foundSymbol SymbolWithLocation

	for _, def := range line {
		matches := loc.Line == def.Loc.Start.Line &&
			loc.Character >= def.Loc.Start.Character &&
			loc.Character <= def.Loc.End.Character

		// Make sure we find the match with the narrowest bounds
		narrower := !alreadyFound || def.Loc.Size() < foundSymbol.Loc.Size()

		if matches && narrower {
			alreadyFound = true
			foundSymbol = def
		}
	}

	if alreadyFound {
		return foundSymbol.Symbol
	}
	return nil
}
