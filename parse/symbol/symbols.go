package symbol

import (
	"fmt"
	"java-mini-ls-go/javaparser"
	"java-mini-ls-go/parse"
	"java-mini-ls-go/parse/loc"
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

var ScopeTypesToCodeSymboltypes = map[parse.ScopeType]CodeSymbolType{
	parse.ScopeTypeAnnotationType:         CodeSymbolClass,
	parse.ScopeTypeClass:                  CodeSymbolClass,
	parse.ScopeTypeConstructor:            CodeSymbolConstructor,
	parse.ScopeTypeEnum:                   CodeSymbolEnum,
	parse.ScopeTypeGenericConstructor:     CodeSymbolConstructor,
	parse.ScopeTypeGenericInterfaceMethod: CodeSymbolMethod,
	parse.ScopeTypeGenericMethod:          CodeSymbolMethod,
	parse.ScopeTypeInterface:              CodeSymbolInterface,
	parse.ScopeTypeInterfaceMethod:        CodeSymbolMethod,
	parse.ScopeTypeMethod:                 CodeSymbolMethod,
	parse.ScopeTypeRecord:                 CodeSymbolClass,
}

// CodeSymbol represents a single symbol inside a source file, whether it's a class, a method, a field, a variable, etc.
type CodeSymbol struct {
	// Name is the name of the symbol
	Name string
	// Type is the type of the symbol
	Type CodeSymbolType
	// Detail is an optional detail about the symbol - method signature, field type/default value, etc.
	Detail string
	// Bounds is the location in the code of this symbol
	Bounds loc.Bounds
	// Children is a list of all CodeSymbols nested under this one
	Children []*CodeSymbol
}

func NewCodeSymbol(name string, ttype CodeSymbolType, fromRule antlr.ParserRuleContext) *CodeSymbol {
	startToken := fromRule.GetStart()
	stopToken := fromRule.GetStop()

	return &CodeSymbol{
		Name:   name,
		Type:   ttype,
		Detail: "",
		Bounds: loc.Bounds{
			Start: loc.FileLocation{Line: startToken.GetLine(), Column: startToken.GetColumn()},
			End:   loc.FileLocation{Line: stopToken.GetLine(), Column: stopToken.GetColumn()},
		},
		Children: make([]*CodeSymbol, 0),
	}
}

func (cs *CodeSymbol) AddChild(symbol *CodeSymbol) {
	cs.Children = append(cs.Children, symbol)
}

func (cs *CodeSymbol) stringRecursive(recursionLevel int) string {
	var childrenStr string
	if cs.Children != nil && len(cs.Children) > 1 {
		indent := strings.Repeat("\t", recursionLevel)
		childrenStr = indent + strings.Join(util.Map(cs.Children, func(cs *CodeSymbol) string {
			return cs.stringRecursive(recursionLevel + 1)
		}), ",\n"+indent)
		childrenStr = fmt.Sprintf(",Children=[\n%s\n]", childrenStr)
	}
	return fmt.Sprintf("(Name=%s,Type=%s,Loc=%s%s)", cs.Name, CodeSymbolTypeNames[cs.Type], cs.Bounds.String(), childrenStr)
}

func (cs *CodeSymbol) String() string {
	return cs.stringRecursive(1)
}

// FindSymbols is the entrypoint for finding topLevelSymbols in a source file
func FindSymbols(tree antlr.Tree) []*CodeSymbol {
	visitor := &symbolVisitor{
		BaseJavaParserListener: javaparser.BaseJavaParserListener{},
		scopeTracker:           parse.NewScopeTracker(),
		topLevelSymbols:        make([]*CodeSymbol, 0),
		symbolStack:            util.NewStack[*CodeSymbol](),
	}
	antlr.ParseTreeWalkerDefault.Walk(visitor, tree)
	return visitor.topLevelSymbols
}

type symbolVisitor struct {
	javaparser.BaseJavaParserListener
	scopeTracker    *parse.ScopeTracker
	topLevelSymbols []*CodeSymbol
	symbolStack     util.Stack[*CodeSymbol]
}

var _ javaparser.JavaParserListener = (*symbolVisitor)(nil)

func (s *symbolVisitor) addSymbol(symbol *CodeSymbol) {
	topSymbol := s.symbolStack.Top()
	if topSymbol != nil {
		topSymbol.AddChild(symbol)
	} else {
		s.topLevelSymbols = append(s.topLevelSymbols, symbol)
	}
}

func (s *symbolVisitor) addNewSymbol(name string, ttype CodeSymbolType, ctx antlr.ParserRuleContext) {
	symbol := NewCodeSymbol(name, ttype, ctx)
	s.addSymbol(symbol)
}

func (s *symbolVisitor) EnterEveryRule(ctx antlr.ParserRuleContext) {
	newScope := s.scopeTracker.CheckEnterScope(ctx)
	if newScope != nil {
		symbolForNewScope := NewCodeSymbol(newScope.Name, ScopeTypesToCodeSymboltypes[newScope.Type], ctx)
		s.addSymbol(symbolForNewScope)
		s.symbolStack.Push(symbolForNewScope)
	}
}

func (s *symbolVisitor) ExitEveryRule(ctx antlr.ParserRuleContext) {
	oldScope := s.scopeTracker.CheckExitScope(ctx)
	if oldScope != nil {
		s.symbolStack.Pop()
	}
}

// EnterPackageDeclaration is called when production packageDeclaration is entered.
func (s *symbolVisitor) EnterPackageDeclaration(ctx *javaparser.PackageDeclarationContext) {
	s.addNewSymbol(ctx.QualifiedName().GetText(), CodeSymbolPackage, ctx)
}

// EnterEnumConstant is called when production enumConstant is entered.
func (s *symbolVisitor) EnterEnumConstant(ctx *javaparser.EnumConstantContext) {
	s.addNewSymbol(ctx.Identifier().GetText(), CodeSymbolEnumMember, ctx)
}

// EnterConstantDeclarator is called when production constantDeclarator is entered.
func (s *symbolVisitor) EnterConstantDeclarator(ctx *javaparser.ConstantDeclaratorContext) {
	s.addNewSymbol(ctx.Identifier().GetText(), CodeSymbolConstant, ctx)
}

// EnterVariableDeclaratorId is called when production variableDeclarator is entered.
func (s *symbolVisitor) EnterVariableDeclaratorId(ctx *javaparser.VariableDeclaratorIdContext) {
	s.addNewSymbol(ctx.Identifier().GetText(), CodeSymbolVariable, ctx)
}

// EnterModuleDeclaration is called when production moduleDeclaration is entered.
func (s *symbolVisitor) EnterModuleDeclaration(ctx *javaparser.ModuleDeclarationContext) {
	s.addNewSymbol(ctx.QualifiedName().GetText(), CodeSymbolPackage, ctx)
}

// EnterCatchClause is called when production catchClause is entered.
func (s *symbolVisitor) EnterCatchClause(ctx *javaparser.CatchClauseContext) {
	s.addNewSymbol(ctx.Identifier().GetText(), CodeSymbolVariable, ctx)
}

// EnterLambdaParameters is called when production lambdaParameters is entered.
func (s *symbolVisitor) EnterLambdaParameters(ctx *javaparser.LambdaParametersContext) {
	for _, ident := range ctx.AllIdentifier() {
		s.addNewSymbol(ident.GetText(), CodeSymbolVariable, ctx)
	}
}
