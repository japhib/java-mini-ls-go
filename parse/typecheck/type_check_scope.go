package typecheck

import "java-mini-ls-go/parse"

type TypeCheckingScope struct {
	Locals   map[string]SymbolWithDefUsages
	Location parse.Bounds
	Parent   *TypeCheckingScope
	Children []TypeCheckingScope
}

func newTypeCheckingScope(parent *TypeCheckingScope, bounds parse.Bounds) TypeCheckingScope {
	newScope := TypeCheckingScope{
		Locals:   make(map[string]SymbolWithDefUsages),
		Children: []TypeCheckingScope{},
		Location: bounds,
	}

	if parent != nil {
		newScope.Parent = parent
		parent.Children = append(parent.Children, newScope)
	}

	return newScope
}

func (tcs *TypeCheckingScope) addLocal(name string, ttype *parse.JavaType, bounds parse.Bounds, fileURI string) {
	tcs.Locals[name] = SymbolWithDefUsages{
		SymbolName: name,
		SymbolType: ttype,
		Definition: parse.CodeLocation{
			FileUri: fileURI,
			Loc:     bounds,
		},
		Usages: make([]parse.CodeLocation, 0),
	}
}

func (tcs *TypeCheckingScope) Contains(location parse.FileLocation) bool {
	withinLines := location.Line >= tcs.Location.Start.Line && location.Line <= tcs.Location.End.Line
	if !withinLines {
		return false
	}

	// one-line scope
	if tcs.Location.Start.Line == tcs.Location.End.Line {
		return location.Column >= tcs.Location.Start.Column && location.Column <= tcs.Location.End.Column
	}

	// First Line
	if location.Line == tcs.Location.Start.Line {
		return location.Column >= tcs.Location.Start.Column
	}

	// last Line
	if location.Line == tcs.Location.End.Line {
		return location.Column <= tcs.Location.End.Column
	}

	// it's on a middle line, columns don't matter
	return true
}

func (tcs *TypeCheckingScope) LookupScopeFor(location parse.FileLocation) *TypeCheckingScope {
	for _, childScope := range tcs.Children {
		// If any of the children match, recurse into that child.
		// This way we find the narrowest scope that matches a given code location.
		if childScope.Contains(location) {
			return childScope.LookupScopeFor(location)
		}
	}

	// If we've gotten this far, there are no children, or none of the children
	// match. So the current scope is the best match.
	return tcs
}
