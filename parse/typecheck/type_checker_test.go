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
					Line:      3,
					Character: 16,
				},
				End: loc.FileLocation{
					Line:      3,
					Character: 20,
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
					Line:      4,
					Character: 22,
				},
				End: loc.FileLocation{
					Line:      4,
					Character: 23,
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
					Line:      5,
					Character: 6,
				},
				End: loc.FileLocation{
					Line:      5,
					Character: 7,
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

func TestCheckTypes_LocalVarUsage(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
public class MainClass {
	public void add() {
		var a = "hi";
		String b = a;
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{}, typeErrors)

	defUsages := typeCheckResult.DefUsagesLookup

	assertSymbol := func(line, col int, kind typ.JavaSymbolKind, name string) {
		result := defUsages.Lookup(loc.FileLocation{
			Line:      line,
			Character: col,
		})
		assert.NotNilf(t, result, "Nil result for lookup at %d:%d", line, col)
		if result != nil {
			assert.Equal(t, kind, result.Kind(), "Wrong type for lookup at %d:%d", line, col)
			assert.Equal(t, name, result.ShortName(), "Wrong ShortName for lookup at %d:%d", line, col)
		}
	}

	assertSymbol(2, 18, typ.JavaSymbolType, "MainClass")
	assertSymbol(3, 13, typ.JavaSymbolMethod, "add()")
	assertSymbol(4, 6, typ.JavaSymbolLocal, "a")
	assertSymbol(5, 9, typ.JavaSymbolLocal, "b")
}

func TestCheckTypes_FieldUsage(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
public class MainClass {
	int a = 1;
	public int getA() {
		return a;
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{}, typeErrors)

	// Get definition: from `return a` to `int a = 1`
	defUsages := typeCheckResult.DefUsagesLookup
	defResult := defUsages.Lookup(loc.FileLocation{Line: 5, Character: 9})
	assert.NotNil(t, defResult)
	if defResult != nil {
		assert.Equal(t, typ.JavaSymbolField, defResult.Kind())
		assert.Equal(t, "a", defResult.ShortName())

		assert.Equal(t, &loc.CodeLocation{
			FileUri: "type_checker_test",
			Loc: loc.Bounds{
				Start: loc.FileLocation{Line: 3, Character: 5},
				End:   loc.FileLocation{Line: 3, Character: 6},
			},
		}, defResult.GetDefinition())
	}

	// Get references: from `int a = 1` to `return a`
	refResult := defUsages.Lookup(loc.FileLocation{Line: 3, Character: 5})
	assert.NotNil(t, refResult)
	if refResult != nil {
		assert.Equal(t, typ.JavaSymbolField, refResult.Kind())
		assert.Equal(t, "a", refResult.ShortName())

		assert.Equal(t, []loc.CodeLocation{{
			FileUri: "type_checker_test",
			Loc: loc.Bounds{
				Start: loc.FileLocation{Line: 5, Character: 9},
				End:   loc.FileLocation{Line: 5, Character: 10},
			},
		}}, refResult.GetUsages())
	}
}

func TestCheckTypes_MethodUsage(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
public class MainClass {
	public void main() {
		int gotten = get1();
	}

	public int get1() {
		return 1;
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{}, typeErrors)

	// Get definition
	defUsages := typeCheckResult.DefUsagesLookup
	defResult := defUsages.Lookup(loc.FileLocation{Line: 4, Character: 15})
	assert.NotNil(t, defResult)
	if defResult != nil {
		assert.Equal(t, typ.JavaSymbolMethod, defResult.Kind())
		assert.Equal(t, "get1", defResult.ShortName())

		assert.Equal(t, &loc.CodeLocation{
			FileUri: "type_checker_test",
			Loc: loc.Bounds{
				Start: loc.FileLocation{Line: 7, Character: 12},
				End:   loc.FileLocation{Line: 7, Character: 16},
			},
		}, defResult.GetDefinition())
	}

	// Get references
	refResult := defUsages.Lookup(loc.FileLocation{Line: 7, Character: 13})
	assert.NotNil(t, refResult)
	if refResult != nil {
		assert.Equal(t, typ.JavaSymbolMethod, refResult.Kind())
		assert.Equal(t, "get1", refResult.ShortName())

		assert.Equal(t, []loc.CodeLocation{{
			FileUri: "type_checker_test",
			Loc: loc.Bounds{
				Start: loc.FileLocation{Line: 4, Character: 15},
				End:   loc.FileLocation{Line: 4, Character: 19},
			},
		}}, refResult.GetUsages())
	}
}

func TestCheckTypes_MethodReturnType(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
public class MainClass {
	public void main() {
		int gotten = getS();
	}

	public String getS() {
		return "hi";
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{{
		Loc: loc.Bounds{
			Start: loc.FileLocation{Line: 4, Character: 15},
			End:   loc.FileLocation{Line: 4, Character: 19},
		},
		Message: "Type mismatch: cannot convert from String to int",
	}}, typeErrors)
}
