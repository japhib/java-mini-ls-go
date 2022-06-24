package parse

import (
	"github.com/antlr/antlr4/runtime/Go/antlr"
	"java-mini-ls-go/javaparser"
)

// Traverse the given parse tree and gather all class, method, field, etc. declarations

type TypeGatheringScope struct {
	Type ScopeType

	// Name of the scope, might end up being unused
	Name string
	// Loc is the full location of the scope. Can be null if it's the file scope.
	Loc *Bounds

	Children []*TypeGatheringScope
	Parent   *TypeGatheringScope
}

type TypeGatheringScopeCreator struct{}

var _ ScopeCreator[TypeGatheringScope] = (*TypeGatheringScopeCreator)(nil)

func (t *TypeGatheringScopeCreator) ShouldCreateScope(ruleType int) bool {
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

func (t *TypeGatheringScopeCreator) CreateScope(parent *TypeGatheringScope, ctx antlr.ParserRuleContext) *TypeGatheringScope {
	loc := ParserRuleContextToBounds(ctx)
	scope := &TypeGatheringScope{
		Loc:      &loc,
		Children: make([]*TypeGatheringScope, 0),
		Parent:   parent,
	}

	parent.Children = append(parent.Children, scope)

	switch ctx.GetRuleIndex() {
	case javaparser.JavaParserRULE_classDeclaration:
		scope.Type = ScopeTypeClassDeclaration
	case javaparser.JavaParserRULE_methodDeclaration:
		scope.Type = ScopeTypeMethodDeclaration
	case javaparser.JavaParserRULE_genericMethodDeclaration:
		scope.Type = ScopeTypeGenericMethodDeclaration
	case javaparser.JavaParserRULE_constructorDeclaration:
		scope.Type = ScopeTypeConstructorDeclaration
	case javaparser.JavaParserRULE_genericConstructorDeclaration:
		scope.Type = ScopeTypeGenericConstructorDeclaration
	case javaparser.JavaParserRULE_interfaceDeclaration:
		scope.Type = ScopeTypeInterfaceDeclaration
	case javaparser.JavaParserRULE_enumDeclaration:
		scope.Type = ScopeTypeEnumDeclaration
	case javaparser.JavaParserRULE_annotationTypeDeclaration:
		scope.Type = ScopeTypeAnnotationTypeDeclaration
	case javaparser.JavaParserRULE_recordDeclaration:
		scope.Type = ScopeTypeRecordDeclaration
	}

	return scope
}

type typeGatheringVisitor struct {
	*javaparser.BaseJavaParserListener
	scopeTracker *ScopeTracker[TypeGatheringScope]
}
