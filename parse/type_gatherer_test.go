package parse

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGatherTypes_Basic(t *testing.T) {
	tree, errors := Parse(`
package stuff;

import somepkg.Thing;
import somepkg.nestedpkg.Nibble;

// declares a type
public class Main {
   // declares a method on type Main
   // Note I took off the '[]' from the typical 'String[] args' to make testing easier 
   public static void main(String args) {
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

	strType := &JavaType{
		Name: "String",
	}

	builtins := TypeMap{
		"String": strType,
	}

	types := GatherTypes(tree, builtins)

	expectedTypes := TypeMap{
		"Main": {
			Name:    "Main",
			Package: "stuff",
			Methods: map[string]*JavaMethod{
				"main": {
					Name: "main",
					// void -> nil
					ReturnType: nil,
					Params: []*JavaParameter{
						{
							Name:      "args",
							Type:      strType,
							IsVarargs: false,
						},
					},
				},
			},
			Constructors: []*JavaConstructor{},
			Fields:       map[string]*JavaField{},
			Type:         JavaTypeClass,
			Extends:      []*JavaType{},
			Implements:   []*JavaType{},
		},
	}

	assert.Equal(t, expectedTypes, types)
}

func TestGatherTypes_NestedClass(t *testing.T) {
	tree, errors := Parse(`
class MyClass {
	public String name;
	public int asdf;
	private Nested n;

	public MyClass(int f) {
		char a = 'a';
	}

	public int DoSomething() {
		int declaredVar = 5;
		return declaredVar;
	}

	class Nested {
		public int nestedInt;
	}
}`)
	assert.Equal(t, 0, len(errors))

	strType := &JavaType{
		Name: "String",
	}
	intType := &JavaType{
		Name: "int",
	}

	builtins := TypeMap{
		"String": strType,
		"int":    intType,
	}

	types := GatherTypes(tree, builtins)

	nestedType := &JavaType{
		Name: "Nested",
		Type: JavaTypeClass,
		Fields: map[string]*JavaField{
			"nestedInt": {
				Name: "nestedInt",
				Type: intType,
			},
		},
		Constructors: []*JavaConstructor{},
		Methods:      map[string]*JavaMethod{},
		Extends:      []*JavaType{},
		Implements:   []*JavaType{},
	}

	expectedTypes := TypeMap{
		"MyClass": {
			Name: "MyClass",
			Fields: map[string]*JavaField{
				"name": {
					Name: "name",
					Type: strType,
				},
				"asdf": {
					Name: "asdf",
					Type: intType,
				},
				"n": {
					Name: "n",
					Type: nestedType,
				},
			},
			Constructors: []*JavaConstructor{
				{
					Arguments: []*JavaParameter{
						{
							Name: "f",
							Type: intType,
						},
					},
				},
			},
			Methods: map[string]*JavaMethod{
				"DoSomething": {
					Name:       "DoSomething",
					ReturnType: intType,
					Params:     []*JavaParameter{},
				},
			},
			Type:       JavaTypeClass,
			Extends:    []*JavaType{},
			Implements: []*JavaType{},
		},
		"Nested": nestedType,
	}

	assert.Equal(t, expectedTypes, types)
}

func TestGatherTypes_Enum(t *testing.T) {
	tree, errors := Parse(`
enum MyEnum {
  First("f"), Second("s"), Third("t");

  private String value;

  MyEnum(String v) {
    this.value = v;
  }

  public String getValue() {
    return this.value;
  }
}`)
	assert.Equal(t, 0, len(errors))

	strType := &JavaType{Name: "String"}
	builtins := TypeMap{"String": strType}
	types := GatherTypes(tree, builtins)

	expectedTypes := TypeMap{
		"MyEnum": &JavaType{
			Name: "MyEnum",
			Type: JavaTypeEnum,
			Constructors: []*JavaConstructor{
				{
					Arguments: []*JavaParameter{
						{
							Name: "v",
							Type: strType,
						},
					},
				},
			},
			Fields: map[string]*JavaField{
				"value": {
					Name: "value",
					Type: strType,
				},
			},
			Methods: map[string]*JavaMethod{
				"getValue": {
					Name:       "getValue",
					ReturnType: strType,
					Params:     []*JavaParameter{},
				},
			},
		},
	}

	assert.Equal(t, expectedTypes, types)
}
