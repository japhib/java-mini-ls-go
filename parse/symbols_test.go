package parse

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type ExpectedSymbol struct {
	Name     string
	Type     CodeSymbolType
	Children []ExpectedSymbol
}

func CheckEquals(t *testing.T, prefix string, expected []ExpectedSymbol, actual []*CodeSymbol) {
	if prefix == "" {
		prefix = "root"
	}

	assert.Equal(t, len(expected), len(actual), "Length not equal for %s", prefix)

	if len(expected) == len(actual) {
		for i, exp := range expected {
			assert.Equal(t, exp.Name, actual[i].Name, "Name not equal for %s[%d]", prefix, i)
			assert.Equal(t, exp.Type, actual[i].Type, "Type not equal for %s[%d]", prefix, i)

			if exp.Children != nil {
				// recurse to children
				CheckEquals(t, prefix+"."+exp.Name, exp.Children, actual[i].Children)
			}
		}
	}
}

func TestFindSymbols_Simple(t *testing.T) {
	tree, errors := Parse(`
class MyClass {
	public int asdf;
}
`)
	assert.Equal(t, 0, len(errors))
	symbols := FindSymbols(tree)

	expected := []ExpectedSymbol{
		{"MyClass", CodeSymbolClass, []ExpectedSymbol{
			{"asdf", CodeSymbolVariable, nil},
		}},
	}

	CheckEquals(t, "", expected, symbols)
}

func TestFindSymbols_NestedClass(t *testing.T) {
	tree, errors := Parse(`
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

	expected := []ExpectedSymbol{
		{"MyClass", CodeSymbolClass, []ExpectedSymbol{
			{"name", CodeSymbolVariable, nil},
			{"asdf", CodeSymbolVariable, nil},
			{"MyClass", CodeSymbolConstructor, []ExpectedSymbol{
				{"a", CodeSymbolVariable, nil},
			}},
			{"DoSomething", CodeSymbolMethod, []ExpectedSymbol{
				{"declaredVar", CodeSymbolVariable, nil},
			}},
			{"Nested", CodeSymbolClass, []ExpectedSymbol{
				{"nestedInt", CodeSymbolVariable, nil},
			}},
		}},
		{"MyEnum", CodeSymbolEnum, []ExpectedSymbol{
			{"First", CodeSymbolEnumMember, nil},
			{"Second", CodeSymbolEnumMember, nil},
			{"Third", CodeSymbolEnumMember, nil},
		}},
	}

	CheckEquals(t, "", expected, symbols)
}
