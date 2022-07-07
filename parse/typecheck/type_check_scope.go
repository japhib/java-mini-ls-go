package typecheck

import (
	"java-mini-ls-go/parse/loc"
	"java-mini-ls-go/parse/typ"
)

type TypeCheckingScope struct {
	Symbol   typ.JavaSymbol
	Locals   map[string]*typ.JavaLocal
	Location loc.Bounds
	Parent   *TypeCheckingScope
	Children []*TypeCheckingScope
}

func newTypeCheckingScope(symbol typ.JavaSymbol, parent *TypeCheckingScope, bounds loc.Bounds) *TypeCheckingScope {
	newScope := &TypeCheckingScope{
		Symbol:   symbol,
		Locals:   make(map[string]*typ.JavaLocal),
		Location: bounds,
		Parent:   parent,
		Children: []*TypeCheckingScope{},
	}

	if parent != nil {
		newScope.Parent = parent
		parent.Children = append(parent.Children, newScope)
	}

	return newScope
}

func (tcs *TypeCheckingScope) addLocal(local *typ.JavaLocal) {
	tcs.Locals[local.Name] = local
}

func (tcs *TypeCheckingScope) Contains(location loc.FileLocation) bool {
	withinLines := location.Line >= tcs.Location.Start.Line && location.Line <= tcs.Location.End.Line
	if !withinLines {
		return false
	}

	// one-line scope
	if tcs.Location.Start.Line == tcs.Location.End.Line {
		return location.Character >= tcs.Location.Start.Character && location.Character <= tcs.Location.End.Character
	}

	// First Line
	if location.Line == tcs.Location.Start.Line {
		return location.Character >= tcs.Location.Start.Character
	}

	// last Line
	if location.Line == tcs.Location.End.Line {
		return location.Character <= tcs.Location.End.Character
	}

	// it's on a middle line, columns don't matter
	return true
}

func (tcs *TypeCheckingScope) LookupScopeFor(location loc.FileLocation) *TypeCheckingScope {
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

// AllSymbols returns a list of all symbols that are relevant to the current scope, including
// symbols from parent scopes.
//
// Used by LSP auto-completion provider.
func (tcs *TypeCheckingScope) AllSymbols() []typ.JavaSymbol {
	var symbols []typ.JavaSymbol

	for _, l := range tcs.Locals {
		symbols = append(symbols, l)
	}

	if ttype, ok := tcs.Symbol.(*typ.JavaType); ok {
		for _, f := range ttype.Fields {
			symbols = append(symbols, f)
		}
		for _, m := range ttype.Methods {
			symbols = append(symbols, m)
		}
	}

	if tcs.Parent != nil {
		symbols = append(symbols, tcs.Parent.AllSymbols()...)
	}

	return symbols
}
