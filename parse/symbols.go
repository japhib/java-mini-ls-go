package parse

import (
	"fmt"
	"java-mini-ls-go/javaparser"
	"java-mini-ls-go/util"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

type CodeSymbolType int

// The various types of CodeSymbolType values
const (
	CodeSymbolClass       CodeSymbolType = iota
	CodeSymbolConstant    CodeSymbolType = iota
	CodeSymbolConstructor CodeSymbolType = iota
	CodeSymbolEnum        CodeSymbolType = iota
	CodeSymbolEnumMember  CodeSymbolType = iota
	CodeSymbolInterface   CodeSymbolType = iota
	CodeSymbolMethod      CodeSymbolType = iota
	CodeSymbolPackage     CodeSymbolType = iota
	CodeSymbolVariable    CodeSymbolType = iota
)

var CodeSymbolTypeNames = map[CodeSymbolType]string{
	CodeSymbolClass:       "Class",
	CodeSymbolConstant:    "Constant",
	CodeSymbolConstructor: "Constructor",
	CodeSymbolEnum:        "Enum",
	CodeSymbolEnumMember:  "EnumMember",
	CodeSymbolInterface:   "Interface",
	CodeSymbolMethod:      "Method",
	CodeSymbolPackage:     "Package",
	CodeSymbolVariable:    "Variable",
}

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

func (cs *CodeSymbol) stringRecursive(recursionLevel int) string {
	if cs.Children != nil && len(cs.Children) > 1 {
		indent := strings.Repeat("\t", recursionLevel)
		childrenStr := indent + strings.Join(util.Map(cs.Children, func(cs *CodeSymbol) string {
			return cs.stringRecursive(recursionLevel + 1)
		}), ",\n"+indent)
		return fmt.Sprintf("(Name=%s,Type=%s,Children=[\n%s\n])", cs.Name, CodeSymbolTypeNames[cs.Type], childrenStr)
	}
	return fmt.Sprintf("(Name=%s,Type=%s)", cs.Name, CodeSymbolTypeNames[cs.Type])
}

func (cs *CodeSymbol) String() string {
	return cs.stringRecursive(1)
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
	case javaparser.JavaParserRULE_interfaceMethodDeclaration:
		ret.Type = CodeSymbolMethod
		body := ctx.(*javaparser.InterfaceMethodDeclarationContext).InterfaceCommonBodyDeclaration().(*javaparser.InterfaceCommonBodyDeclarationContext)
		ret.Name = body.Identifier().GetText()
	case javaparser.JavaParserRULE_genericInterfaceMethodDeclaration:
		ret.Type = CodeSymbolMethod
		body := ctx.(*javaparser.InterfaceMethodDeclarationContext).InterfaceCommonBodyDeclaration().(*javaparser.InterfaceCommonBodyDeclarationContext)
		ret.Name = body.Identifier().GetText()
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

func (s *symbolVisitor) addSymbolToPreviousScope(symbol *CodeSymbol) {
	scopeCount := s.scopeTracker.GetScopeCount()
	if scopeCount <= 1 {
		// Add to top level
		s.symbols = append(s.symbols, symbol)
	} else {
		// Add to the previous top scope
		secondToTop := s.scopeTracker.GetTopScopeMinus(1)
		secondToTop.Children = append(secondToTop.Children, symbol)
	}
}

func (s *symbolVisitor) addSymbol(symbol *CodeSymbol) {
	topScope := s.scopeTracker.GetTopScope()
	if topScope != nil {
		topScope.Children = append(topScope.Children, symbol)
	} else {
		s.symbols = append(s.symbols, symbol)
	}
}

func (s *symbolVisitor) addNewSymbol(name string, ttype CodeSymbolType) {
	symbol := &CodeSymbol{
		Name: name,
		Type: ttype,
	}
	s.addSymbol(symbol)
}

func (s *symbolVisitor) EnterEveryRule(ctx antlr.ParserRuleContext) {
	if s.scopeTracker.CheckEnterScope(ctx) {
		newScope := s.scopeTracker.GetTopScope()
		s.addSymbolToPreviousScope(newScope)
	}
}

func (s *symbolVisitor) ExitEveryRule(ctx antlr.ParserRuleContext) {
	s.scopeTracker.CheckExitScope(ctx)
}

// EnterPackageDeclaration is called when production packageDeclaration is entered.
func (s *symbolVisitor) EnterPackageDeclaration(ctx *javaparser.PackageDeclarationContext) {
	s.addNewSymbol(ctx.QualifiedName().GetText(), CodeSymbolPackage)
}

// EnterEnumConstant is called when production enumConstant is entered.
func (s *symbolVisitor) EnterEnumConstant(ctx *javaparser.EnumConstantContext) {
	s.addNewSymbol(ctx.Identifier().GetText(), CodeSymbolEnumMember)
}

// EnterConstantDeclarator is called when production constantDeclarator is entered.
func (s *symbolVisitor) EnterConstantDeclarator(ctx *javaparser.ConstantDeclaratorContext) {
	s.addNewSymbol(ctx.Identifier().GetText(), CodeSymbolConstant)
}

// EnterVariableDeclaratorId is called when production variableDeclarator is entered.
func (s *symbolVisitor) EnterVariableDeclaratorId(ctx *javaparser.VariableDeclaratorIdContext) {
	s.addNewSymbol(ctx.Identifier().GetText(), CodeSymbolVariable)
}

// EnterModuleDeclaration is called when production moduleDeclaration is entered.
func (s *symbolVisitor) EnterModuleDeclaration(ctx *javaparser.ModuleDeclarationContext) {
	s.addNewSymbol(ctx.QualifiedName().GetText(), CodeSymbolPackage)
}

// EnterCatchClause is called when production catchClause is entered.
func (s *symbolVisitor) EnterCatchClause(ctx *javaparser.CatchClauseContext) {
	s.addNewSymbol(ctx.Identifier().GetText(), CodeSymbolVariable)
}

// EnterLambdaParameters is called when production lambdaParameters is entered.
func (s *symbolVisitor) EnterLambdaParameters(ctx *javaparser.LambdaParametersContext) {
	for _, ident := range ctx.AllIdentifier() {
		s.addNewSymbol(ident.GetText(), CodeSymbolVariable)
	}
}
