package typecheck

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"java-mini-ls-go/parse"
	"java-mini-ls-go/parse/loc"
	"java-mini-ls-go/parse/typ"
	"testing"
)

func parseAndTypeCheck(t *testing.T, code string) TypeCheckResult {
	// Make sure to load built-in types
	_, err := typ.LoadBuiltinTypes()
	if err != nil {
		t.Fatalf("Error loading builtin types: %s", err.Error())
	}

	tree, parseErrors := parse.Parse(code)
	assert.Equal(t, 0, len(parseErrors))

	strType := &typ.JavaType{Name: "String"}
	objType := &typ.JavaType{Name: "Object"}
	builtins := typ.TypeMap{"String": strType, "Object": objType}
	typ.AddPrimitiveTypes(builtins)

	return CheckTypes(zaptest.NewLogger(t), tree, "type_checker_test", builtins)
}

func TestCheckTypes_Addition(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
public class MainClass {
	public int add() {
		return 2 + 1;
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{}, typeErrors)
}

func TestCheckTypes_Params(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
public class MainClass {
	public int addOne(int a) {
		return a + 1;
	}

	public double add(int a, double b) {
		return a + b;
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{}, typeErrors)
}

func TestCheckTypes_ClassVars(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
public class MainClass {
	int aA;
	static double bB;

	public int addA(int a) {
		return a + aA;
	}

	public double addB(double b) {
		return b + bB;
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{}, typeErrors)
}

func TestCheckTypes_ClassVarsInSuperclass(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
public class Parent {
	int aA;
	static double bB;
}

public class MainClass extends Parent {
	public int addA(int a) {
		// references instance var in superclass
		return a + aA;
	}

	public double addB(double b) {
		// references static var in superclass
		return b + bB;
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{}, typeErrors)
}

func TestCheckTypes_AddsLocalVars(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
public class MainClass {
	public int add() {
		int a = 1;
		return a + 1;
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{}, typeErrors)
}

func TestCheckTypes_CheckLocalVars(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
public class MainClass {
	int a = 1, b = 2, c = 3;

	public void add() {
		int d = 4, e = 5, f = 6;
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{}, typeErrors)
}

func TestCheckTypes_CheckFieldVars_Errors(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
public class MainClass {
	int a = 1, b = "hi";
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{
		{
			Loc: loc.Bounds{
				Start: loc.FileLocation{
					Line:   3,
					Column: 16,
				},
				End: loc.FileLocation{
					Line:   3,
					Column: 20,
				},
			},
			Message: "Type mismatch: cannot convert from String to int",
		},
	}, typeErrors)
}

func TestCheckTypes_CheckLocalVars_Errors(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
public class MainClass {
	public void add() {
		String c = "f", a = 5;
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{
		{
			Loc: loc.Bounds{
				Start: loc.FileLocation{
					Line:   4,
					Column: 22,
				},
				End: loc.FileLocation{
					Line:   4,
					Column: 23,
				},
			},
			Message: "Type mismatch: cannot convert from int to String",
		},
	}, typeErrors)
}

func TestCheckTypes_CheckLocalVars_ErrorRedefined(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
public class MainClass {
	public void add() {
		int a;
		int a = 1;
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{
		{
			Loc: loc.Bounds{
				Start: loc.FileLocation{
					Line:   5,
					Column: 6,
				},
				End: loc.FileLocation{
					Line:   5,
					Column: 7,
				},
			},
			Message: "Variable a is already defined in method add",
		},
	}, typeErrors)
}

func TestCheckTypes_CheckVarDecl(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
public class MainClass {
	public void add() {
		var a = "hi";
		String b = a;
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{}, typeErrors)
}
