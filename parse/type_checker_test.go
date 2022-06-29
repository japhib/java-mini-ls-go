package parse

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func parseAndTypeCheck(t *testing.T, code string) []TypeError {
	tree, parseErrors := Parse(code)
	assert.Equal(t, 0, len(parseErrors))

	strType := &JavaType{Name: "String"}
	objType := &JavaType{Name: "Object"}
	builtins := TypeMap{"String": strType, "Object": objType}
	addPrimitiveTypes(builtins)
	userTypes := GatherTypes(tree, builtins)

	return CheckTypes(tree, userTypes, builtins)
}

func TestCheckTypes_Addition(t *testing.T) {
	typeErrors := parseAndTypeCheck(t, `
public class MainClass {
	public int add() {
		return 2 + 1;
	}
}`)
	assert.Equal(t, []TypeError{}, typeErrors)
}

func TestCheckTypes_Params(t *testing.T) {
	typeErrors := parseAndTypeCheck(t, `
public class MainClass {
	public int addOne(int a) {
		return a + 1;
	}

	public double add(int a, double b) {
		return a + b;
	}
}`)
	assert.Equal(t, []TypeError{}, typeErrors)
}

func TestCheckTypes_ClassVars(t *testing.T) {
	typeErrors := parseAndTypeCheck(t, `
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
	assert.Equal(t, []TypeError{}, typeErrors)
}

func TestCheckTypes_ClassVarsInSuperclass(t *testing.T) {
	typeErrors := parseAndTypeCheck(t, `
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
	assert.Equal(t, []TypeError{}, typeErrors)
}

func TestCheckTypes_AddsLocalVars(t *testing.T) {
	typeErrors := parseAndTypeCheck(t, `
public class MainClass {
	public int add() {
		int a = 1;
		return a + 1;
	}
}`)
	assert.Equal(t, []TypeError{}, typeErrors)
}

func TestCheckTypes_CheckLocalVars(t *testing.T) {
	typeErrors := parseAndTypeCheck(t, `
public class MainClass {
	int a = 1, b = 2, c = 3;

	public void add() {
		int d = 4, e = 5, f = 6;
	}
}`)
	assert.Equal(t, []TypeError{}, typeErrors)
}

func TestCheckTypes_CheckFieldVars_Errors(t *testing.T) {
	typeErrors := parseAndTypeCheck(t, `
public class MainClass {
	int a = 1, b = "hi";
}`)
	assert.Equal(t, []TypeError{
		{
			Loc: Bounds{
				Start: FileLocation{
					Line:   3,
					Column: 16,
				},
				End: FileLocation{
					Line:   3,
					Column: 16,
				},
			},
			Message: "Type mismatch: cannot convert from String to int",
		},
	}, typeErrors)
}

func TestCheckTypes_CheckLocalVars_Errors(t *testing.T) {
	typeErrors := parseAndTypeCheck(t, `
public class MainClass {
	public void add() {
		String c = "f", a = 5;
	}
}`)
	assert.Equal(t, []TypeError{
		{
			Loc: Bounds{
				Start: FileLocation{
					Line:   4,
					Column: 22,
				},
				End: FileLocation{
					Line:   4,
					Column: 22,
				},
			},
			Message: "Type mismatch: cannot convert from int to String",
		},
	}, typeErrors)
}

func TestCheckTypes_CheckLocalVars_ErrorRedefined(t *testing.T) {
	typeErrors := parseAndTypeCheck(t, `
public class MainClass {
	public void add() {
		int a;
		int a = 1;
	}
}`)
	assert.Equal(t, []TypeError{
		{
			Loc: Bounds{
				Start: FileLocation{
					Line:   4,
					Column: 22,
				},
				End: FileLocation{
					Line:   4,
					Column: 22,
				},
			},
			Message: "Local variable a was defined previously",
		},
	}, typeErrors)
}
