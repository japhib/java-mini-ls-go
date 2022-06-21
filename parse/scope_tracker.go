package parse

import (
	"github.com/antlr/antlr4/runtime/Go/antlr"
)

type ScopeCreator[TScope any] interface {
	ShouldCreateScope(ruleType int) bool
	CreateScope(ctx antlr.RuleContext) *TScope
}

type ScopeTracker[TScope any] struct {
	// A stack of scopes
	ScopeStack []*TScope
	Creator    ScopeCreator[TScope]
}

func NewScopeTracker[TScope any](scopeCreator ScopeCreator[TScope]) *ScopeTracker[TScope] {
	return &ScopeTracker[TScope]{
		ScopeStack: []*TScope{},
		Creator:    scopeCreator,
	}
}

func (sv *ScopeTracker[TScope]) CheckEnterScope(ctx antlr.ParserRuleContext) bool {
	if sv.Creator.ShouldCreateScope(ctx.GetRuleIndex()) {
		// Create new scope and add to stack
		sv.ScopeStack = append(sv.ScopeStack, sv.Creator.CreateScope(ctx))
		return true
	}
	return false
}

func (sv *ScopeTracker[TScope]) CheckExitScope(ctx antlr.ParserRuleContext) bool {
	if sv.Creator.ShouldCreateScope(ctx.GetRuleIndex()) {
		// Pop top scope from the stack
		sv.ScopeStack = sv.ScopeStack[:len(sv.ScopeStack)-1]
		return true
	}
	return false
}

func (sv *ScopeTracker[TScope]) GetTopScope() *TScope {
	if len(sv.ScopeStack) == 0 {
		return nil
	}

	return sv.ScopeStack[len(sv.ScopeStack)-1]
}

func (sv *ScopeTracker[TScope]) GetTopScopeMinus(offsetFromTop int) *TScope {
	if len(sv.ScopeStack)-offsetFromTop <= 0 {
		return nil
	}

	return sv.ScopeStack[len(sv.ScopeStack)-1-offsetFromTop]
}

func (sv *ScopeTracker[TScope]) GetScopeCount() int {
	return len(sv.ScopeStack)
}
