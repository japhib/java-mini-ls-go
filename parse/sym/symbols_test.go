package sym

import (
	"github.com/stretchr/testify/assert"
	"java-mini-ls-go/parse"
	"java-mini-ls-go/parse/loc"
	"testing"
)

// Used for zeroing out the "bounds" of the given code symbols in case you
// want to test equality but not worry about bounds
func zeroBounds(symbols []*CodeSymbol) {
	for _, symbol := range symbols {
		var b loc.Bounds
		symbol.Bounds = b

		if symbol.Children != nil && len(symbol.Children) > 0 {
			zeroBounds(symbol.Children)
		} else {
			symbol.Children = nil
		}
	}
}

func TestFindSymbols_Simple(t *testing.T) {
	tree, errors := parse.Parse(`
class MyClass {
	public int asdf;
}
`)
	assert.Equal(t, 0, len(errors))
	symbols := FindSymbols(tree)
	zeroBounds(symbols)

	expected := []*CodeSymbol{
		{Name: "MyClass", Type: CodeSymbolClass, Children: []*CodeSymbol{
			{Name: "asdf", Type: CodeSymbolVariable, Children: nil},
		}},
	}

	assert.Equal(t, expected, symbols)
}

func TestFindSymbols_WithMain(t *testing.T) {
	tree, errors := parse.Parse(`
package java;

import somepkg.Thing;
import somepkg.nestedpkg.Nibble;

// declares a type
public class Main {
   // declares a method on type Main
   public static void main(String[] args) {
       // function call "println" on type of "System.out" using string arg
       System.out.println("Hi there");

       Thing thing = new Thing("pub!", 3);
       System.out.println("the value of pubfield is " + thing.pubfield);
       System.out.println("the value of privfield is " + thing.getPrivField());

       Nibble nibble = new Nibble("asdf", 5);
       System.out.println("nibble: " + nibble);
   }
}`)

	assert.Equal(t, 0, len(errors))
	symbols := FindSymbols(tree)
	zeroBounds(symbols)

	expected := []*CodeSymbol{
		{Name: "java", Type: CodeSymbolPackage, Children: nil},
		{Name: "Main", Type: CodeSymbolClass, Children: []*CodeSymbol{
			{Name: "main", Type: CodeSymbolMethod, Children: []*CodeSymbol{
				{Name: "args", Type: CodeSymbolVariable, Children: nil},
				{Name: "thing", Type: CodeSymbolVariable, Children: nil},
				{Name: "nibble", Type: CodeSymbolVariable, Children: nil},
			}},
		}},
	}

	assert.Equal(t, expected, symbols)
}

func TestFindSymbols_NestedClass(t *testing.T) {
	tree, errors := parse.Parse(`
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
	assert.Equal(t, 0, len(errors))
	symbols := FindSymbols(tree)
	zeroBounds(symbols)

	expected := []*CodeSymbol{
		{Name: "MyClass", Type: CodeSymbolClass, Children: []*CodeSymbol{
			{Name: "name", Type: CodeSymbolVariable, Children: nil},
			{Name: "asdf", Type: CodeSymbolVariable, Children: nil},
			{Name: "MyClass", Type: CodeSymbolConstructor, Children: []*CodeSymbol{
				{Name: "a", Type: CodeSymbolVariable, Children: nil},
			}},
			{Name: "DoSomething", Type: CodeSymbolMethod, Children: []*CodeSymbol{
				{Name: "declaredVar", Type: CodeSymbolVariable, Children: nil},
			}},
			{Name: "Nested", Type: CodeSymbolClass, Children: []*CodeSymbol{
				{Name: "nestedInt", Type: CodeSymbolVariable, Children: nil},
			}},
		}},
		{Name: "MyEnum", Type: CodeSymbolEnum, Children: []*CodeSymbol{
			{Name: "First", Type: CodeSymbolEnumMember, Children: nil},
			{Name: "Second", Type: CodeSymbolEnumMember, Children: nil},
			{Name: "Third", Type: CodeSymbolEnumMember, Children: nil},
		}},
	}

	assert.Equal(t, expected, symbols)
}
