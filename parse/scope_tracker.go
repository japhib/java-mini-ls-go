package parse

import (
	"github.com/antlr/antlr4/runtime/Go/antlr"
)

type ScopeCreator[TScope any] interface {
	ShouldCreateScope(ruleType int) bool
	CreateScope(ctx antlr.RuleContext) TScope
}

type ScopeTracker[TScope any] struct {
	// A stack of scopes
	ScopeStack []TScope
	Creator    ScopeCreator[TScope]
}

func NewScopeTracker[TScope any](scopeCreator ScopeCreator[TScope]) *ScopeTracker[TScope] {
	return &ScopeTracker[TScope]{
		ScopeStack: []TScope{},
		Creator:    scopeCreator,
	}
}

func (sv *ScopeTracker[TScope]) CheckEnterScope(ctx antlr.ParserRuleContext) {
	if sv.Creator.ShouldCreateScope(ctx.GetRuleIndex()) {
		// Create new scope and add to stack
		sv.ScopeStack = append(sv.ScopeStack, sv.Creator.CreateScope(ctx))
	}
}

func (sv *ScopeTracker[TScope]) CheckExitScope(ctx antlr.ParserRuleContext) {
	if sv.Creator.ShouldCreateScope(ctx.GetRuleIndex()) {
		// Pop top scope from the stack
		sv.ScopeStack = sv.ScopeStack[:len(sv.ScopeStack)-1]
	}
}

func (sv *ScopeTracker[TScope]) GetTopScope() TScope {
	return sv.ScopeStack[len(sv.ScopeStack)-1]
}
