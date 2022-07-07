package typecheck

import (
	"github.com/stretchr/testify/assert"
	"java-mini-ls-go/parse/typ"
	"testing"
)

func TestTypeCheckingScope_Basic(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
public class MainClass {
	public String add() {
		int a;
		String b;
		return b + a;
	}
}`)
	assert.Equal(t, []TypeError{}, typeCheckResult.TypeErrors)

	rootScope := typeCheckResult.RootScope

	assert.Equal(t, nil, rootScope.Symbol)
	assert.Equal(t, (*TypeCheckingScope)(nil), rootScope.Parent)
	assert.Equal(t, 0, len(rootScope.Locals))
	assert.Equal(t, 1, len(rootScope.Children))

	mainClassScope := rootScope.Children[0]

	assert.Equal(t, typ.JavaSymbolType, mainClassScope.Symbol.Kind())
	assert.Equal(t, "MainClass", mainClassScope.Symbol.ShortName())
	assert.Equal(t, rootScope, mainClassScope.Parent)
	assert.Equal(t, 0, len(mainClassScope.Locals))
	assert.Equal(t, 1, len(mainClassScope.Children))

	addMethodScope := mainClassScope.Children[0]

	assert.Equal(t, typ.JavaSymbolMethod, addMethodScope.Symbol.Kind())
	assert.Equal(t, "add", addMethodScope.Symbol.ShortName())
	assert.Equal(t, mainClassScope, addMethodScope.Parent)
	assert.Equal(t, 2, len(addMethodScope.Locals))
	assert.Equal(t, 0, len(addMethodScope.Children))
}
