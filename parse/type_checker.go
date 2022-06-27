package parse

type TypeCheckingScope struct {
	Type ScopeType

	// Name of the scope, might end up being unused
	Name string
	// Loc is the full location of the scope. Can be null if it's the file scope.
	Loc *Bounds

	Children []*TypeCheckingScope
	Parent   *TypeCheckingScope
}
