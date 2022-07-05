package parse

import (
	"java-mini-ls-go/javaparser"
	"java-mini-ls-go/parse/loc"
	"java-mini-ls-go/util"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"golang.org/x/exp/slices"
)

type ScopeType int

const (
	ScopeTypeUnset                  ScopeType = iota
	ScopeTypeAnnotationType         ScopeType = iota
	ScopeTypeClass                  ScopeType = iota
	ScopeTypeConstructor            ScopeType = iota
	ScopeTypeEnum                   ScopeType = iota
	ScopeTypeGenericConstructor     ScopeType = iota
	ScopeTypeGenericInterfaceMethod ScopeType = iota
	ScopeTypeGenericMethod          ScopeType = iota
	ScopeTypeInterface              ScopeType = iota
	ScopeTypeInterfaceMethod        ScopeType = iota
	ScopeTypeMethod                 ScopeType = iota
	ScopeTypeRecord                 ScopeType = iota
)

var classTypes = []ScopeType{
	ScopeTypeAnnotationType,
	ScopeTypeClass,
	ScopeTypeEnum,
	ScopeTypeInterface,
	ScopeTypeRecord,
}

var methodTypes = []ScopeType{
	ScopeTypeConstructor,
	ScopeTypeGenericConstructor,
	ScopeTypeGenericInterfaceMethod,
	ScopeTypeGenericMethod,
	ScopeTypeInterfaceMethod,
	ScopeTypeMethod,
}

func (st ScopeType) IsClassType() bool {
	return slices.Contains(classTypes, st)
}

func (st ScopeType) IsMethodType() bool {
	return slices.Contains(methodTypes, st)
}

type Scope struct {
	// Name is the name of the scope
	Name string
	// Type is the type of the scope
	Type ScopeType
	// Bounds is the location in the code of this scope
	Bounds loc.Bounds
	// Parent is the outer scope this one is nested under
	Parent *Scope
	// Children is a list of all scopes nested under this one
	Children []*Scope
}

type ScopeTracker struct {
	// A stack of scopes
	ScopeStack util.Stack[*Scope]
}

func NewScopeTracker() *ScopeTracker {
	return &ScopeTracker{
		ScopeStack: util.NewStack[*Scope](),
	}
}

func (st *ScopeTracker) CheckEnterScope(ctx antlr.ParserRuleContext) *Scope {
	if st.shouldCreateScope(ctx.GetRuleIndex()) {
		// Create new scope and add to stack
		newScope := st.createScope(st.ScopeStack.Top(), ctx)
		st.ScopeStack.Push(newScope)
		return newScope
	}
	return nil
}

func (st *ScopeTracker) CheckExitScope(ctx antlr.ParserRuleContext) *Scope {
	if st.shouldCreateScope(ctx.GetRuleIndex()) {
		// Pop top scope from the stack
		return st.ScopeStack.Pop()
	}
	return nil
}

func (st *ScopeTracker) CurrScopeName() string {
	scopeNames := make([]string, 0, st.ScopeStack.Size())
	for i := 0; i < st.ScopeStack.Size(); i++ {
		scopeNames = append(scopeNames, st.ScopeStack.At(i).Name)
	}
	return strings.Join(scopeNames, ".")
}

func (st *ScopeTracker) shouldCreateScope(ruleType int) bool {
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

func (st *ScopeTracker) createScope(parent *Scope, ctx antlr.ParserRuleContext) *Scope {
	ret := &Scope{
		Name:     "",
		Type:     ScopeTypeUnset,
		Bounds:   loc.Bounds{},
		Parent:   parent,
		Children: make([]*Scope, 0),
	}

	switch ctx.GetRuleIndex() {
	case javaparser.JavaParserRULE_classDeclaration:
		ret.Type = ScopeTypeClass
		subCtx := ctx.(*javaparser.ClassDeclarationContext).Identifier()
		ret.Name, ret.Bounds = nameAndBoundsForCtx(subCtx)
	case javaparser.JavaParserRULE_methodDeclaration:
		ret.Type = ScopeTypeMethod
		subCtx := ctx.(*javaparser.MethodDeclarationContext).Identifier()
		ret.Name, ret.Bounds = nameAndBoundsForCtx(subCtx)
	case javaparser.JavaParserRULE_genericMethodDeclaration:
		ret.Type = ScopeTypeGenericMethod
		subCtx := ctx.(*javaparser.GenericMethodDeclarationContext).MethodDeclaration().(*javaparser.MethodDeclarationContext).Identifier()
		ret.Name, ret.Bounds = nameAndBoundsForCtx(subCtx)
	case javaparser.JavaParserRULE_interfaceMethodDeclaration:
		ret.Type = ScopeTypeInterfaceMethod
		subCtx := ctx.(*javaparser.InterfaceMethodDeclarationContext).InterfaceCommonBodyDeclaration().(*javaparser.InterfaceCommonBodyDeclarationContext).Identifier()
		ret.Name, ret.Bounds = nameAndBoundsForCtx(subCtx)
	case javaparser.JavaParserRULE_genericInterfaceMethodDeclaration:
		ret.Type = ScopeTypeGenericInterfaceMethod
		subCtx := ctx.(*javaparser.InterfaceMethodDeclarationContext).InterfaceCommonBodyDeclaration().(*javaparser.InterfaceCommonBodyDeclarationContext).Identifier()
		ret.Name, ret.Bounds = nameAndBoundsForCtx(subCtx)
	case javaparser.JavaParserRULE_constructorDeclaration:
		ret.Type = ScopeTypeConstructor
		subCtx := ctx.(*javaparser.ConstructorDeclarationContext).Identifier()
		ret.Name, ret.Bounds = nameAndBoundsForCtx(subCtx)
	case javaparser.JavaParserRULE_genericConstructorDeclaration:
		ret.Type = ScopeTypeGenericConstructor
		subCtx := ctx.(*javaparser.GenericConstructorDeclarationContext).ConstructorDeclaration().(*javaparser.ConstructorDeclarationContext).Identifier()
		ret.Name, ret.Bounds = nameAndBoundsForCtx(subCtx)
	case javaparser.JavaParserRULE_interfaceDeclaration:
		ret.Type = ScopeTypeInterface
		subCtx := ctx.(*javaparser.InterfaceDeclarationContext).Identifier()
		ret.Name, ret.Bounds = nameAndBoundsForCtx(subCtx)
	case javaparser.JavaParserRULE_enumDeclaration:
		ret.Type = ScopeTypeEnum
		subCtx := ctx.(*javaparser.EnumDeclarationContext).Identifier()
		ret.Name, ret.Bounds = nameAndBoundsForCtx(subCtx)
	case javaparser.JavaParserRULE_annotationTypeDeclaration:
		ret.Type = ScopeTypeAnnotationType
		subCtx := ctx.(*javaparser.AnnotationTypeDeclarationContext).Identifier()
		ret.Name, ret.Bounds = nameAndBoundsForCtx(subCtx)
	case javaparser.JavaParserRULE_recordDeclaration:
		ret.Type = ScopeTypeRecord
		subCtx := ctx.(*javaparser.RecordDeclarationContext).Identifier()
		ret.Name, ret.Bounds = nameAndBoundsForCtx(subCtx)
	}

	return ret
}

func nameAndBoundsForCtx(ident javaparser.IIdentifierContext) (string, loc.Bounds) {
	return ident.GetText(), loc.ParserRuleContextToBounds(ident)
}
