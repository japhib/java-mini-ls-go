package typecheck

import "java-mini-ls-go/parse"

// SymbolWithDefUsages keeps tracks of the definition and all the usages of a symbol.
type SymbolWithDefUsages struct {
	// TODO SymbolName is insufficient to uniquely identify a constructor or method since they can be overloaded.
	// Should use fully qualified name - {package}.{class}.{method}.{args}
	SymbolName string
	SymbolType *parse.JavaType
	Definition parse.CodeLocation
	Usages     []parse.CodeLocation
}

func NewSymbolWithDefUsages(name string, ttype *parse.JavaType, definition parse.CodeLocation) *SymbolWithDefUsages {
	return &SymbolWithDefUsages{
		SymbolName: name,
		SymbolType: ttype,
		Definition: definition,
		Usages:     []parse.CodeLocation{},
	}
}

// DefinitionsUsagesWithLocation is a single definition or usage of a symbol.
// Contains the location of the definition/usage, and a pointer to the SymbolWithDefUsages
// struct which contains the list of all the other usages.
//
// Note that it uses Bounds instead of CodeLocation since it only is for a single file.
type DefinitionsUsagesWithLocation struct {
	Loc       parse.Bounds
	DefUsages *SymbolWithDefUsages
}

// DefinitionsUsagesOnLine is a list of all the identifiers on a line of code in the editor,
// their bounds, and all the SymbolWithDefUsages they point to.
type DefinitionsUsagesOnLine []DefinitionsUsagesWithLocation

// DefinitionsUsagesLookup is a lookup table that helps the language server go from just
// a file URI and code location to figuring out what identifier is being pointed to,
// and finding its corresponding SymbolWithDefUsages struct.
type DefinitionsUsagesLookup struct {
	// DefUsagesByLine is a map of line numbers to the list of DefinitionsUsagesWithLocation on that line.
	DefUsagesByLine map[int]DefinitionsUsagesOnLine

	// DefUsagesByName is a map of symbols *defined* in this file (not just used) by name.
	// TODO might be unused
	DefUsagesByName map[string]*SymbolWithDefUsages
}

func NewDefinitionsUsagesLookup() *DefinitionsUsagesLookup {
	return &DefinitionsUsagesLookup{
		DefUsagesByLine: make(map[int]DefinitionsUsagesOnLine),
		DefUsagesByName: make(map[string]*SymbolWithDefUsages),
	}
}

func (dul *DefinitionsUsagesLookup) GetLine(line int) DefinitionsUsagesOnLine {
	return dul.DefUsagesByLine[line]
}

func (dul *DefinitionsUsagesLookup) NewSymbol(loc parse.Bounds, defUsToAdd *SymbolWithDefUsages) {
	lineNumber := loc.Start.Line
	line := dul.DefUsagesByLine[lineNumber]

	// If a list for that line doesn't exist, create it now
	if line != nil {
		line = make(DefinitionsUsagesOnLine, 0)
		dul.DefUsagesByLine[lineNumber] = line
	}

	// Also make sure there's not already an item with that name/bounds
	for _, defUsages := range line {
		if defUsages.DefUsages.SymbolName == defUsToAdd.SymbolName && defUsages.Loc.Equals(loc) {
			// Name/location match
			// TODO merge usages
			return
		}
	}

	// If we've made it this far, it hasn't been found so we can add it
	line = append(line, DefinitionsUsagesWithLocation{
		DefUsages: defUsToAdd,
		Loc:       loc,
	})

	// Make sure to assign it back in case append had to realloc
	dul.DefUsagesByLine[lineNumber] = line

	// Add to DefUsagesByName as well
	dul.DefUsagesByName[defUsToAdd.SymbolName] = defUsToAdd
}

// Lookup Given a file location, returns the most specific SymbolWithDefUsages instance corresponding
// to that file location, if one exists.
func (dul *DefinitionsUsagesLookup) Lookup(loc parse.FileLocation) *SymbolWithDefUsages {
	line := dul.DefUsagesByLine[loc.Line]
	if line == nil {
		return nil
	}

	var found *DefinitionsUsagesWithLocation = nil

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
		return found.DefUsages
	}
	return nil
}
