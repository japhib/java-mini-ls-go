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
	builtins, err := typ.LoadBuiltinTypes()
	if err != nil {
		t.Fatalf("Error loading builtin types: %s", err.Error())
	}

	tree, parseErrors := parse.Parse(code)
	assert.Equal(t, 0, len(parseErrors))

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
	assertSymbol(3, 13, typ.JavaSymbolMethod, "add")
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
			End:   loc.FileLocation{Line: 4, Character: 21},
		},
		Message: "Type mismatch: cannot convert from String to int",
	}}, typeErrors)
}

func TestCheckTypes_MethodArgumentTypes_Success(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
public class MainClass {
	public void main() {
		var gotten = getS("hi", 1);
	}

	public String getS(String s, int a) {
		return s;
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{}, typeErrors)
}

func TestCheckTypes_MethodArgumentTypes_Error(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
public class MainClass {
	public void main() {
		var gotten = getS(1, 1, 1);
	}

	public String getS(int a, String s, int b) {
		return s;
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{{
		Loc: loc.Bounds{
			Start: loc.FileLocation{
				Line:      4,
				Character: 15,
			},
			End: loc.FileLocation{
				Line:      4,
				Character: 28,
			},
		},
		Message: "Can't use int as type String in function call to getS",
	}}, typeErrors)
}

func TestCheckTypes_MethodArgumentTypes_ErrorNotEnoughArguments(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
public class MainClass {
	public void main() {
		var gotten = getS();
	}

	public String getS(int a) {
		return "s";
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{{
		Loc: loc.Bounds{
			Start: loc.FileLocation{
				Line:      4,
				Character: 15,
			},
			End: loc.FileLocation{
				Line:      4,
				Character: 21,
			},
		},
		Message: "Not enough arguments in function call to getS! Expected 1, got 0",
	}}, typeErrors)
}

func TestCheckTypes_DotOperator_Field(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
class Something {
	int a;
}

class MainClass {
	void main() {
		Something s = new Something();
		var a = s.a;
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{}, typeErrors)

	// Get definition of Something.a
	defUsages := typeCheckResult.DefUsagesLookup
	defResult := defUsages.Lookup(loc.FileLocation{Line: 9, Character: 12})
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
				Start: loc.FileLocation{Line: 9, Character: 12},
				End:   loc.FileLocation{Line: 9, Character: 13},
			},
		}}, refResult.GetUsages())
	}
}

func TestCheckTypes_DotOperator_Field_Error(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
class Something {
	int a;
}

class MainClass {
	void main() {
		Something s = new Something();
		var a = s.b;
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{{
		Loc: loc.Bounds{
			Start: loc.FileLocation{
				Line:      9,
				Character: 12,
			},
			End: loc.FileLocation{
				Line:      9,
				Character: 13,
			},
		},
		Message: "Can't find member named b of type Something",
	}}, typeErrors)
}

func TestCheckTypes_DotOperator_Method(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
class Something {
	static void println() { }
}

class MainClass {
	void main() {
		Something.println();
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{}, typeErrors)

	// Get definition
	defUsages := typeCheckResult.DefUsagesLookup
	defResult := defUsages.Lookup(loc.FileLocation{Line: 8, Character: 13})
	assert.NotNil(t, defResult)
	if defResult != nil {
		assert.Equal(t, typ.JavaSymbolMethod, defResult.Kind())
		assert.Equal(t, "println", defResult.ShortName())

		assert.Equal(t, &loc.CodeLocation{
			FileUri: "type_checker_test",
			Loc: loc.Bounds{
				Start: loc.FileLocation{Line: 3, Character: 13},
				End:   loc.FileLocation{Line: 3, Character: 20},
			},
		}, defResult.GetDefinition())
	}

	// Get references
	refResult := defUsages.Lookup(loc.FileLocation{Line: 3, Character: 13})
	assert.NotNil(t, refResult)
	if refResult != nil {
		assert.Equal(t, typ.JavaSymbolMethod, refResult.Kind())
		assert.Equal(t, "println", refResult.ShortName())

		assert.Equal(t, []loc.CodeLocation{{
			FileUri: "type_checker_test",
			Loc: loc.Bounds{
				Start: loc.FileLocation{Line: 8, Character: 12},
				End:   loc.FileLocation{Line: 8, Character: 19},
			},
		}}, refResult.GetUsages())
	}
}

func TestCheckTypes_DotOperator_Method_Error(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
class MainClass {
	void main() {
		var a = System.in;
		System.out.printlg();
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{{
		Loc: loc.Bounds{
			Start: loc.FileLocation{
				Line:      5,
				Character: 13,
			},
			End: loc.FileLocation{
				Line:      5,
				Character: 20,
			},
		},
		Message: "Can't find member named printlg on type PrintStream",
	}}, typeErrors)
}

func TestCheckTypes_DotOperator_Method_OneArgument(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
class MainClass {
	void main() {
		var a = System.in;
		System.out.println("hi");
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{}, typeErrors)
}

func TestCheckTypes_DotOperator_Method_OneArgument_Error(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
class MainClass {
	void main() {
		var a = System.in;
		System.out.append("not_a_char");
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{{
		Loc: loc.Bounds{
			Start: loc.FileLocation{
				Line:      5,
				Character: 13,
			},
			End: loc.FileLocation{
				Line:      5,
				Character: 33,
			},
		},
		Message: "Can't use String as type char in function call to PrintStream.append",
	}}, typeErrors)
}

func TestCheckTypes_DotOperator_Method_SeveralArguments(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
class HelperClass {
	static void print3things(String a, int b, int c) {
		System.out.println(a + b + c);
	}
}

class MainClass {
	void main() {
		HelperClass.print3things("asdf", 1, 2);
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{}, typeErrors)
}

func TestCheckTypes_DotOperator_Method_SeveralArguments_Error(t *testing.T) {
	typeCheckResult := parseAndTypeCheck(t, `
class HelperClass {
	static void print3things(String a, int b, int c) {
		System.out.println(a + b + c);
	}
}

class MainClass {
	void main() {
		HelperClass.print3things("asdf", 1);
	}
}`)
	typeErrors := typeCheckResult.TypeErrors
	assert.Equal(t, []TypeError{
		{
			Loc: loc.Bounds{
				Start: loc.FileLocation{
					Line:      10,
					Character: 14,
				},
				End: loc.FileLocation{
					Line:      10,
					Character: 37,
				},
			},
			Message: "Can't use String as type int in function call to HelperClass.print3things",
		},
		{
			Loc: loc.Bounds{
				Start: loc.FileLocation{
					Line:      10,
					Character: 14,
				},
				End: loc.FileLocation{
					Line:      10,
					Character: 37,
				},
			},
			Message: "Not enough arguments in function call to HelperClass.print3things! Expected 3, got 2",
		},
	}, typeErrors)
}
