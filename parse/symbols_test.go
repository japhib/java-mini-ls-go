package parse

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFindSymbols(t *testing.T) {
	tree := Parse(`
class MyClass {
	public String name;
	public int asdf;

	public MyClass() {
		char a = 'a';
	}

	public void DoSomething() {
		int declaredVar = 5;
	}

	class Nested {
		public int nestedInt;
	}
}

enum MyEnum {
	First, Second, Third
}
`)
	symbols := FindSymbols(tree)
	fmt.Println(symbols)

	assert.Equal(t, 2, len(symbols))
	assert.Equal(t, "MyClass", symbols[0].Name)
	assert.Equal(t, CodeSymbolClass, symbols[0].Type)
	assert.Equal(t, "MyEnum", symbols[1].Name)
	assert.Equal(t, CodeSymbolEnum, symbols[1].Type)

	myclassChildren := symbols[0].Children
	assert.Equal(t, 3, len(myclassChildren))
	assert.Equal(t, "MyClass", myclassChildren[0].Name)
	assert.Equal(t, CodeSymbolConstructor, myclassChildren[0].Type)
	assert.Equal(t, "DoSomething", myclassChildren[1].Name)
	assert.Equal(t, CodeSymbolMethod, myclassChildren[1].Type)
	assert.Equal(t, "Nested", myclassChildren[2].Name)
	assert.Equal(t, CodeSymbolClass, myclassChildren[2].Type)
}
