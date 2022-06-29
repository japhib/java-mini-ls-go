package parse

import (
	"fmt"
	"golang.org/x/exp/slices"
	"java-mini-ls-go/util"
	"strings"
)

type VisibilityType int

const (
	VisibilityDefault   VisibilityType = iota
	VisibilityPrivate   VisibilityType = iota
	VisibilityPublic    VisibilityType = iota
	VisibilityProtected VisibilityType = iota
)

var VisibilityTypeStrs = map[VisibilityType]string{
	VisibilityDefault:   "<package-public>",
	VisibilityPrivate:   "private",
	VisibilityPublic:    "public",
	VisibilityProtected: "protected",
}

type JavaTypeType int

const (
	JavaTypePrimitive  JavaTypeType = iota
	JavaTypeClass      JavaTypeType = iota
	JavaTypeInterface  JavaTypeType = iota
	JavaTypeEnum       JavaTypeType = iota
	JavaTypeRecord     JavaTypeType = iota
	JavaTypeAnnotation JavaTypeType = iota
)

var JavaTypeTypeStrs = map[JavaTypeType]string{
	JavaTypePrimitive:  "primitive",
	JavaTypeClass:      "class",
	JavaTypeInterface:  "interface",
	JavaTypeEnum:       "enum",
	JavaTypeRecord:     "record",
	JavaTypeAnnotation: "annotation",
}

func getStaticStr(isStatic bool) string {
	if isStatic {
		return "static "
	} else {
		return ""
	}
}

type JavaType struct {
	Name    string
	Package string
	Module  string
	// Note Extends is a slice only because interfaces can extend multiple other interfaces.
	// For classes this will have a maximum of one element.
	Extends      []*JavaType
	Implements   []*JavaType
	Constructors []*JavaConstructor
	Fields       map[string]*JavaField
	// TODO handle method overloads
	Methods    map[string]*JavaMethod
	Visibility VisibilityType
	Type       JavaTypeType
}

func (jt *JavaType) LookupField(name string) *JavaField {
	if field, ok := jt.Fields[name]; ok {
		return field
	}

	// Go to parent class/interfaces and see if any of them have the field
	for _, supertype := range jt.Extends {
		field := supertype.LookupField(name)
		if field != nil {
			return field
		}
	}

	// Not found
	return nil
}

func (jt *JavaType) String() string {
	return fmt.Sprintf("%s %s %s", VisibilityTypeStrs[jt.Visibility], JavaTypeTypeStrs[jt.Type], jt.Name)
}

func (jt *JavaType) AllSuperClasses() []*JavaType {
	supers := []*JavaType{}

	if jt.Extends == nil {
		return supers
	}

	for _, e := range jt.Extends {
		// Append immediate superclass
		supers = append(supers, e)

		// Recursively append superclasses of that superclass
		for _, ee := range e.AllSuperClasses() {
			supers = append(supers, ee)
		}
	}

	return supers
}

// Map of which other primitive/boxed types each primitive type can coerce to
var primitivesCoercions = map[string][]string{
	"byte":    {"Byte", "short", "int", "long", "float", "double"},
	"short":   {"Short", "int", "long", "float", "double"},
	"int":     {"Integer", "long", "float", "double"},
	"long":    {"Long", "float", "double"},
	"float":   {"Float", "double"},
	"double":  {"Double"},
	"char":    {"Character", "int", "long", "float", "double"},
	"boolean": {"Boolean"},
}

// Map of boxed primitives back to their unboxed primitives
var boxedPrimitives = map[string]string{
	"Byte":    "byte",
	"short":   "Short",
	"int":     "Integer",
	"long":    "Long",
	"float":   "Float",
	"double":  "Double",
	"char":    "Character",
	"boolean": "Boolean",
}

// CoercesTo says whether a type can be converted to another type without a type cast.
func (jt *JavaType) CoercesTo(other *JavaType) bool {
	if jt == other {
		return true
	}

	// Any type, including primitives, can be coerced to java.lang.Object
	if other.Name == "Object" {
		return true
	}

	// If it's a primitive type, there are several coercions that can be made automatically
	if jt.Type == JavaTypePrimitive {
		return slices.Contains(primitivesCoercions[jt.Name], other.Name)
	}

	// Boxed primitive types can be converted to the non-boxed primitives
	if jt.Type == JavaTypeClass && other.Type == JavaTypePrimitive {
		if boxed, ok := boxedPrimitives[jt.Name]; ok && other.Name == boxed {
			return true
		}
	}

	// Is it a superclass of this type?
	superclasses := jt.AllSuperClasses()
	for _, super := range superclasses {
		if super == other {
			return true
		}
	}

	// In any other case, it's false
	return false
}

type JavaField struct {
	Name       string
	Type       *JavaType
	Visibility VisibilityType
	IsStatic   bool
	IsFinal    bool
}

func (jf *JavaField) String() string {
	return fmt.Sprintf("%s %s%s %s", VisibilityTypeStrs[jf.Visibility], getStaticStr(jf.IsStatic), jf.Type.Name, jf.Name)
}

type JavaConstructor struct {
	Visibility VisibilityType
	Arguments  []*JavaParameter
}

type JavaMethod struct {
	Name       string
	ReturnType *JavaType
	Params     []*JavaParameter
	Visibility VisibilityType
	IsStatic   bool
}

func (jm *JavaMethod) String() string {
	argStr := ""
	if jm.Params != nil {
		argStr = strings.Join(util.MapToString(jm.Params), ", ")
	}

	return fmt.Sprintf("%s %s%s %s(%s)", VisibilityTypeStrs[jm.Visibility], getStaticStr(jm.IsStatic), jm.ReturnType, jm.Name, argStr)
}

type JavaParameter struct {
	Name      string
	Type      *JavaType
	IsVarargs bool
}

func (ja *JavaParameter) String() string {
	return fmt.Sprintf("%s %s", ja.Type.Name, ja.Name)
}
