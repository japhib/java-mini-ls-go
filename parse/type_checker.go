package parse

type ScopeType int

const (
	ScopeTypeAnnotationTypeDeclaration     ScopeType = iota
	ScopeTypeClassDeclaration              ScopeType = iota
	ScopeTypeConstructorDeclaration        ScopeType = iota
	ScopeTypeEnumDeclaration               ScopeType = iota
	ScopeTypeGenericConstructorDeclaration ScopeType = iota
	ScopeTypeGenericMethodDeclaration      ScopeType = iota
	ScopeTypeInterfaceDeclaration          ScopeType = iota
	ScopeTypeMethodDeclaration             ScopeType = iota
	ScopeTypeRecordDeclaration             ScopeType = iota
)

type TypeCheckingScope struct {
	Type ScopeType

	// Name of the scope, might end up being unused
	Name string
	// Loc is the full location of the scope. Can be null if it's the file scope.
	Loc *Bounds

	Children []*TypeCheckingScope
	Parent   *TypeCheckingScope
}
