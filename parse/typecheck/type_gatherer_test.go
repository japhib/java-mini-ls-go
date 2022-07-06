package typecheck

import (
	"github.com/stretchr/testify/assert"
	"java-mini-ls-go/parse"
	"java-mini-ls-go/parse/typ"
	"testing"
)

func stripOutStuffWeDontWannaTest(types *typ.TypeMap) {
	nilOutCircularRefs(types)
	stripAllsDefsUsages(types)
}

func stripAllsDefsUsages(types *typ.TypeMap) {
	for _, ttype := range types.AllTypes() {
		stripDefsUsages(ttype)
	}
}

func stripDefsUsages(ttype *typ.JavaType) {
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

	strType := &typ.JavaType{
		Name: "String",
	}

	builtins := typ.NewTypeMap()
	builtins.Add(strType)

	types, _ := GatherTypes("testfile", tree, builtins)
	stripOutStuffWeDontWannaTest(types)

	expectedTypes := typ.NewTypeMap()
	expectedTypes.Add(&typ.JavaType{
		Name:    "Main",
		Package: "stuff",
		Methods: []*typ.JavaMethod{
			{
				Name: "main",
				// void -> nil
				ReturnType: nil,
				Params: []*typ.JavaParameter{
					{
						Name:      "args",
						Type:      strType,
						IsVarargs: false,
					},
				},
			},
		},
		Constructors: []*typ.JavaConstructor{},
		Fields:       []*typ.JavaField{},
		Type:         typ.JavaTypeClass,
		Extends:      []*typ.JavaType{},
		Implements:   []*typ.JavaType{},
		Visibility:   typ.VisibilityPublic,
	})

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

	strType := &typ.JavaType{
		Name:       "String",
		Visibility: typ.VisibilityPublic,
	}
	intType := &typ.JavaType{
		Name:       "int",
		Visibility: typ.VisibilityPublic,
	}

	builtins := typ.NewTypeMap()
	builtins.Add(strType)
	builtins.Add(intType)

	types, _ := GatherTypes("testfile", tree, builtins)
	stripOutStuffWeDontWannaTest(types)

	nestedType := &typ.JavaType{
		Name: "Nested",
		Type: typ.JavaTypeClass,
		Fields: []*typ.JavaField{
			{
				Name: "nestedInt",
				Type: intType,
			},
		},
		Constructors: []*typ.JavaConstructor{},
		Methods:      []*typ.JavaMethod{},
		Extends:      []*typ.JavaType{},
		Implements:   []*typ.JavaType{},
		Visibility:   typ.VisibilityPublic,
	}

	expectedTypes := typ.NewTypeMap()
	expectedTypes.Add(&typ.JavaType{
		Name: "MyClass",
		Fields: []*typ.JavaField{
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
		Constructors: []*typ.JavaConstructor{
			{
				Params: []*typ.JavaParameter{
					{
						Name: "f",
						Type: intType,
					},
				},
			},
		},
		Methods: []*typ.JavaMethod{
			{
				Name:       "DoSomething",
				ReturnType: intType,
				Params:     []*typ.JavaParameter{},
			},
		},
		Type:       typ.JavaTypeClass,
		Extends:    []*typ.JavaType{},
		Implements: []*typ.JavaType{},
		Visibility: typ.VisibilityPublic,
	})
	expectedTypes.Add(nestedType)

	assert.Equal(t, expectedTypes, types)
}

// nilOutCircularRefs sets backreferences (ParentType) inside a type to be nil
// so that it's easier to test
func nilOutCircularRefs(types *typ.TypeMap) {
	for _, ttype := range types.AllTypes() {
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

	strType := &typ.JavaType{Name: "String"}
	builtins := typ.NewTypeMap()
	builtins.Add(strType)

	types, _ := GatherTypes("testfile", tree, builtins)
	stripOutStuffWeDontWannaTest(types)

	expectedTypes := typ.NewTypeMap()
	expectedTypes.Add(&typ.JavaType{
		Name: "MyEnum",
		Type: typ.JavaTypeEnum,
		Constructors: []*typ.JavaConstructor{
			{
				Params: []*typ.JavaParameter{
					{
						Name: "v",
						Type: strType,
					},
				},
			},
		},
		Fields: []*typ.JavaField{
			{
				Name: "value",
				Type: strType,
			},
		},
		Methods: []*typ.JavaMethod{
			{
				Name:       "getValue",
				ReturnType: strType,
				Params:     []*typ.JavaParameter{},
			},
		},
		Extends:    []*typ.JavaType{},
		Implements: []*typ.JavaType{},
		Visibility: typ.VisibilityPublic,
	})

	assert.Equal(t, expectedTypes, types)
}
