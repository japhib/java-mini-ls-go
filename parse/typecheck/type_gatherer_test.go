package typecheck

import (
	"github.com/stretchr/testify/assert"
	"java-mini-ls-go/parse"
	"testing"
)

func stripOutStuffWeDontWannaTest(types parse.TypeMap) {
	nilOutCircularRefs(types)
	stripAllsDefsUsages(types)
}

func stripAllsDefsUsages(types parse.TypeMap) {
	for _, ttype := range types {
		stripDefsUsages(ttype)
	}
}

func stripDefsUsages(ttype *parse.JavaType) {
	// TODO test this stuff when the location stuff is more accurate

	ttype.Definition = nil
	ttype.Usages = nil

	for _, c := range ttype.Constructors {
		c.Definition = nil
		c.Usages = nil
	}

	for _, m := range ttype.Methods {
		m.Definition = nil
		m.Usages = nil
	}

	for _, f := range ttype.Fields {
		f.Definition = nil
		f.Usages = nil
	}
}

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

	types := GatherTypes("testfile", tree, builtins)
	stripOutStuffWeDontWannaTest(types)

	expectedTypes := parse.TypeMap{
		"Main": {
			Name:    "Main",
			Package: "stuff",
			Methods: []*parse.JavaMethod{
				{
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
			Fields:       []*parse.JavaField{},
			Type:         parse.JavaTypeClass,
			Extends:      []*parse.JavaType{},
			Implements:   []*parse.JavaType{},
			Visibility:   parse.VisibilityPublic,
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
		Name:       "String",
		Visibility: parse.VisibilityPublic,
	}
	intType := &parse.JavaType{
		Name:       "int",
		Visibility: parse.VisibilityPublic,
	}

	builtins := parse.TypeMap{
		"String": strType,
		"int":    intType,
	}

	types := GatherTypes("testfile", tree, builtins)
	stripOutStuffWeDontWannaTest(types)

	nestedType := &parse.JavaType{
		Name: "Nested",
		Type: parse.JavaTypeClass,
		Fields: []*parse.JavaField{
			{
				Name: "nestedInt",
				Type: intType,
			},
		},
		Constructors: []*parse.JavaConstructor{},
		Methods:      []*parse.JavaMethod{},
		Extends:      []*parse.JavaType{},
		Implements:   []*parse.JavaType{},
		Visibility:   parse.VisibilityPublic,
	}

	expectedTypes := parse.TypeMap{
		"MyClass": {
			Name: "MyClass",
			Fields: []*parse.JavaField{
				{
					Name: "name",
					Type: strType,
				},
				{
					Name: "asdf",
					Type: intType,
				},
				{
					Name: "n",
					Type: nestedType,
				},
			},
			Constructors: []*parse.JavaConstructor{
				{
					Params: []*parse.JavaParameter{
						{
							Name: "f",
							Type: intType,
						},
					},
				},
			},
			Methods: []*parse.JavaMethod{
				{
					Name:       "DoSomething",
					ReturnType: intType,
					Params:     []*parse.JavaParameter{},
				},
			},
			Type:       parse.JavaTypeClass,
			Extends:    []*parse.JavaType{},
			Implements: []*parse.JavaType{},
			Visibility: parse.VisibilityPublic,
		},
		"Nested": nestedType,
	}

	assert.Equal(t, expectedTypes, types)
}

// nilOutCircularRefs sets backreferences (ParentType) inside a type to be nil
// so that it's easier to test
func nilOutCircularRefs(types parse.TypeMap) {
	for _, ttype := range types {
		for _, constructor := range ttype.Constructors {
			constructor.ParentType = nil
		}
		for _, method := range ttype.Methods {
			method.ParentType = nil
		}
		for _, field := range ttype.Fields {
			field.ParentType = nil
		}
	}
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
	types := GatherTypes("testfile", tree, builtins)
	stripOutStuffWeDontWannaTest(types)

	expectedTypes := parse.TypeMap{
		"MyEnum": &parse.JavaType{
			Name: "MyEnum",
			Type: parse.JavaTypeEnum,
			Constructors: []*parse.JavaConstructor{
				{
					Params: []*parse.JavaParameter{
						{
							Name: "v",
							Type: strType,
						},
					},
				},
			},
			Fields: []*parse.JavaField{
				{
					Name: "value",
					Type: strType,
				},
			},
			Methods: []*parse.JavaMethod{
				{
					Name:       "getValue",
					ReturnType: strType,
					Params:     []*parse.JavaParameter{},
				},
			},
			Extends:    []*parse.JavaType{},
			Implements: []*parse.JavaType{},
			Visibility: parse.VisibilityPublic,
		},
	}

	assert.Equal(t, expectedTypes, types)
}
