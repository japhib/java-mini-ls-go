package parse

// SymbolWithDefUsages keeps tracks of the definition and all the usages of a symbol.
type SymbolWithDefUsages struct {
	SymbolName string
	SymbolType *JavaType
	Definition CodeLocation
	Usages     []CodeLocation
}

// DefinitionsUsagesWithLocation is a single definition or usage of a symbol.
// Contains the location of the definition/usage, and a pointer to the SymbolWithDefUsages
// struct which contains the list of all the other usages.
//
// Note that it uses Bounds instead of CodeLocation since it only is for a single file.
type DefinitionsUsagesWithLocation struct {
	Loc       Bounds
	DefUsages *SymbolWithDefUsages
}

// DefinitionsUsagesOnLine is a list of all the identifiers on a line of code in the editor,
// their bounds, and all the SymbolWithDefUsages they point to.
type DefinitionsUsagesOnLine []DefinitionsUsagesWithLocation

// DefinitionsUsagesLookup is a lookup table that helps the language server go from just
// a file URI and code location to figuring out what identifier is being pointed to,
// and finding its corresponding SymbolWithDefUsages struct.
type DefinitionsUsagesLookup struct {
	// LookupTable is a map of line numbers to the list of DefinitionsUsagesWithLocation on that line.
	LookupTable map[int]DefinitionsUsagesOnLine
}

func (dul *DefinitionsUsagesLookup) GetLine(line int) DefinitionsUsagesOnLine {
	return dul.LookupTable[line]
}

func (dul *DefinitionsUsagesLookup) Add(loc Bounds, defUsToAdd *SymbolWithDefUsages) {
	lineNumber := loc.Start.Line
	line := dul.LookupTable[lineNumber]

	// If a list for that line doesn't exist, create it now
	if line != nil {
		line = make(DefinitionsUsagesOnLine, 0)
		dul.LookupTable[lineNumber] = line
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
	dul.LookupTable[lineNumber] = line
}

// Lookup Given a file location, returns the most specific SymbolWithDefUsages instance corresponding
// to that file location, if one exists.
func (dul *DefinitionsUsagesLookup) Lookup(loc FileLocation) *SymbolWithDefUsages {
	line := dul.LookupTable[loc.Line]
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
