package parse

import (
	"java-mini-ls-go/javaparser"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

type CodeSymbolType int

// The various types of CodeSymbolType values
const (
	CodeSymbolClass         CodeSymbolType = iota
	CodeSymbolConstant      CodeSymbolType = iota
	CodeSymbolConstructor   CodeSymbolType = iota
	CodeSymbolEnum          CodeSymbolType = iota
	CodeSymbolEnumMember    CodeSymbolType = iota
	CodeSymbolField         CodeSymbolType = iota
	CodeSymbolInterface     CodeSymbolType = iota
	CodeSymbolMethod        CodeSymbolType = iota
	CodeSymbolPackage       CodeSymbolType = iota
	CodeSymbolTypeParameter CodeSymbolType = iota
	CodeSymbolVariable      CodeSymbolType = iota
)

// CodeSymbol represents a single symbol inside a source file, whether it's a class, a method, a field, a variable, etc.
type CodeSymbol struct {
	// Name is the name of the symbol
	Name string
	// Type is the type of the symbol
	Type CodeSymbolType
	// Detail is an optional detail about the symbol - method signature, field type/default value, etc.
	Detail string
	// Location is the location in the code of this symbol
	Bounds Bounds
	// Children is a list of all CodeSymbols nested under this one
	Children []*CodeSymbol
}

func FindSymbols(tree antlr.Tree) []*CodeSymbol {
	visitor := &symbolVisitor{
		scopeTracker: NewScopeTracker[CodeSymbol](&symbolScopeCreator{}),
		symbols:      make([]*CodeSymbol, 0),
	}
	antlr.ParseTreeWalkerDefault.Walk(visitor, tree)
	return visitor.symbols
}

type symbolScopeCreator struct{}

var _ ScopeCreator[CodeSymbol] = (*symbolScopeCreator)(nil)

func (sc *symbolScopeCreator) ShouldCreateScope(ruleType int) bool {
	switch ruleType {
	case javaparser.JavaParserRULE_classDeclaration:
		return true
	case javaparser.JavaParserRULE_methodDeclaration:
		return true
	case javaparser.JavaParserRULE_genericMethodDeclaration:
		return true
	case javaparser.JavaParserRULE_constructorDeclaration:
		return true
	case javaparser.JavaParserRULE_genericConstructorDeclaration:
		return true
	case javaparser.JavaParserRULE_interfaceDeclaration:
		return true
	case javaparser.JavaParserRULE_enumDeclaration:
		return true
	case javaparser.JavaParserRULE_annotationTypeDeclaration:
		return true
	case javaparser.JavaParserRULE_recordDeclaration:
		return true
	}
	return false
}

func (sc *symbolScopeCreator) CreateScope(ctx antlr.RuleContext) *CodeSymbol {
	ctx.GetChildren()

	ret := &CodeSymbol{
		Children: make([]*CodeSymbol, 0),
	}

	switch ctx.GetRuleIndex() {
	case javaparser.JavaParserRULE_classDeclaration:
		ret.Type = CodeSymbolClass
		ret.Name = ctx.(*javaparser.ClassDeclarationContext).Identifier().GetText()
		//ret.Bounds = GetClassDeclBounds(ctx.(*javaparser.ClassDeclarationContext))
	case javaparser.JavaParserRULE_methodDeclaration:
		ret.Type = CodeSymbolMethod
		ret.Name = ctx.(*javaparser.MethodDeclarationContext).Identifier().GetText()
	case javaparser.JavaParserRULE_genericMethodDeclaration:
		ret.Type = CodeSymbolMethod
		ret.Name = ctx.(*javaparser.GenericMethodDeclarationContext).MethodDeclaration().(*javaparser.MethodDeclarationContext).Identifier().GetText()
	case javaparser.JavaParserRULE_constructorDeclaration:
		ret.Type = CodeSymbolConstructor
		ret.Name = ctx.(*javaparser.ConstructorDeclarationContext).Identifier().GetText()
	case javaparser.JavaParserRULE_genericConstructorDeclaration:
		ret.Type = CodeSymbolConstructor
		ret.Name = ctx.(*javaparser.GenericConstructorDeclarationContext).ConstructorDeclaration().(*javaparser.ConstructorDeclarationContext).Identifier().GetText()
	case javaparser.JavaParserRULE_interfaceDeclaration:
		ret.Type = CodeSymbolInterface
		ret.Name = ctx.(*javaparser.InterfaceDeclarationContext).Identifier().GetText()
	case javaparser.JavaParserRULE_enumDeclaration:
		ret.Type = CodeSymbolEnum
		ret.Name = ctx.(*javaparser.EnumDeclarationContext).Identifier().GetText()
	case javaparser.JavaParserRULE_annotationTypeDeclaration:
		ret.Type = CodeSymbolClass
		ret.Name = ctx.(*javaparser.AnnotationTypeDeclarationContext).Identifier().GetText()
	case javaparser.JavaParserRULE_recordDeclaration:
		ret.Type = CodeSymbolClass
		ret.Name = ctx.(*javaparser.RecordDeclarationContext).Identifier().GetText()
	}

	return ret
}

type symbolVisitor struct {
	javaparser.BaseJavaParserListener
	scopeTracker *ScopeTracker[CodeSymbol]
	symbols      []*CodeSymbol
}

var _ javaparser.JavaParserListener = (*symbolVisitor)(nil)

func (s *symbolVisitor) EnterEveryRule(ctx antlr.ParserRuleContext) {
	if s.scopeTracker.CheckEnterScope(ctx) {
		justAdded := s.scopeTracker.GetTopScope()

		scopeCount := s.scopeTracker.GetScopeCount()
		if scopeCount == 1 {
			// Add to top level
			s.symbols = append(s.symbols, justAdded)
		} else {
			// Add to the previous top scope
			secondToTop := s.scopeTracker.GetTopScopeMinus(1)
			secondToTop.Children = append(secondToTop.Children, justAdded)
		}
	}
}

func (s *symbolVisitor) ExitEveryRule(ctx antlr.ParserRuleContext) {
	s.scopeTracker.CheckExitScope(ctx)
}
