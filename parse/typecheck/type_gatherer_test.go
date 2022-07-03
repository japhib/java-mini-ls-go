package typecheck

import (
	"github.com/stretchr/testify/assert"
	"java-mini-ls-go/parse"
	"testing"
)

func TestGatherTypes_Basic(t *testing.T) {
	tree, errors := parse.Parse(`
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

	strType := &parse.JavaType{
		Name: "String",
	}

	builtins := parse.TypeMap{
		"String": strType,
	}

	types := GatherTypes(tree, builtins)

	expectedTypes := parse.TypeMap{
		"Main": {
			Name:    "Main",
			Package: "stuff",
			Methods: map[string]*parse.JavaMethod{
				"main": {
					Name: "main",
					// void -> nil
					ReturnType: nil,
					Params: []*parse.JavaParameter{
						{
							Name:      "args",
							Type:      strType,
							IsVarargs: false,
						},
					},
				},
			},
			Constructors: []*parse.JavaConstructor{},
			Fields:       map[string]*parse.JavaField{},
			Type:         parse.JavaTypeClass,
			Extends:      []*parse.JavaType{},
			Implements:   []*parse.JavaType{},
		},
	}

	assert.Equal(t, expectedTypes, types)
}

func TestGatherTypes_NestedClass(t *testing.T) {
	tree, errors := parse.Parse(`
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

	strType := &parse.JavaType{
		Name: "String",
	}
	intType := &parse.JavaType{
		Name: "int",
	}

	builtins := parse.TypeMap{
		"String": strType,
		"int":    intType,
	}

	types := GatherTypes(tree, builtins)

	nestedType := &parse.JavaType{
		Name: "Nested",
		Type: parse.JavaTypeClass,
		Fields: map[string]*parse.JavaField{
			"nestedInt": {
				Name: "nestedInt",
				Type: intType,
			},
		},
		Constructors: []*parse.JavaConstructor{},
		Methods:      map[string]*parse.JavaMethod{},
		Extends:      []*parse.JavaType{},
		Implements:   []*parse.JavaType{},
	}

	expectedTypes := parse.TypeMap{
		"MyClass": {
			Name: "MyClass",
			Fields: map[string]*parse.JavaField{
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
			Constructors: []*parse.JavaConstructor{
				{
					Arguments: []*parse.JavaParameter{
						{
							Name: "f",
							Type: intType,
						},
					},
				},
			},
			Methods: map[string]*parse.JavaMethod{
				"DoSomething": {
					Name:       "DoSomething",
					ReturnType: intType,
					Params:     []*parse.JavaParameter{},
				},
			},
			Type:       parse.JavaTypeClass,
			Extends:    []*parse.JavaType{},
			Implements: []*parse.JavaType{},
		},
		"Nested": nestedType,
	}

	assert.Equal(t, expectedTypes, types)
}

func TestGatherTypes_Enum(t *testing.T) {
	tree, errors := parse.Parse(`
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

	strType := &parse.JavaType{Name: "String"}
	builtins := parse.TypeMap{"String": strType}
	types := GatherTypes(tree, builtins)

	expectedTypes := parse.TypeMap{
		"MyEnum": &parse.JavaType{
			Name: "MyEnum",
			Type: parse.JavaTypeEnum,
			Constructors: []*parse.JavaConstructor{
				{
					Arguments: []*parse.JavaParameter{
						{
							Name: "v",
							Type: strType,
						},
					},
				},
			},
			Fields: map[string]*parse.JavaField{
				"value": {
					Name: "value",
					Type: strType,
				},
			},
			Methods: map[string]*parse.JavaMethod{
				"getValue": {
					Name:       "getValue",
					ReturnType: strType,
					Params:     []*parse.JavaParameter{},
				},
			},
		},
	}

	assert.Equal(t, expectedTypes, types)
}
