package parse

import (
	"java-mini-ls-go/javaparser"
)

type CodeSymbolType int

// The various types of CodeSymbolType values
const (
	SymbolClass = iota
	SymbolConstant
	SymbolConstructor
	SymbolEnum
	SymbolEnumMember
	SymbolField
	SymbolInterface
	SymbolMethod
	SymbolPackage
	SymbolTypeParameter
	SymbolVariable
)

// CodeSymbol represents a single symbol inside a source file, whether it's a class, a method, a field, a variable, etc.
type CodeSymbol struct {
	// Name is the name of the symbol
	Name string
	// Type is the type of the symbol
	Type CodeSymbolType
	// Detail is an optional detail about the symbol - method signature, field type/default value, etc.
	Detail string
}

type symbolWalker struct {
	*javaparser.BaseJavaParserListener
}

func (l *symbolWalker) EnterClassDeclaration(ctx *javaparser.ClassDeclarationContext) {

}
