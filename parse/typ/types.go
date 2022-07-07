package typ

import (
	"fmt"
	"java-mini-ls-go/parse/loc"
	"java-mini-ls-go/util"
	"strings"

	"golang.org/x/exp/slices"
)

type JavaSymbolKind int

const (
	JavaSymbolType        JavaSymbolKind = iota
	JavaSymbolConstructor JavaSymbolKind = iota
	JavaSymbolMethod      JavaSymbolKind = iota
	JavaSymbolField       JavaSymbolKind = iota
	JavaSymbolLocal       JavaSymbolKind = iota
)

type JavaSymbol interface {
	// Kind returns what type of symbol this is
	Kind() JavaSymbolKind

	// PackageName returns the package name for this symbol
	PackageName() string

	// ShortName returns the short name for this symbol
	ShortName() string

	// FullName returns the fully-qualified name for this symbol.
	// However much is applicable of:
	// {package}.{class}.{method_or_field}.{args}
	FullName() string

	// GetVisibility returns the visibility of this symbol
	GetVisibility() VisibilityType

	// GetDefinition returns the location in code where this symbol is defined.
	// May be nil for built-in or library types.
	GetDefinition() *loc.CodeLocation

	// GetUsages returns a list of all usages of this symbol.
	GetUsages() []loc.CodeLocation

	// AddUsage adds a usage of this symbol in the given location.
	AddUsage(loc.CodeLocation)

	// GetType returns the type of this symbol, using a special LSP type for classes & methods
	GetType() *JavaType
}

type VisibilityType int

const (
	VisibilityDefault   VisibilityType = iota
	VisibilityPrivate   VisibilityType = iota
	VisibilityPublic    VisibilityType = iota
	VisibilityProtected VisibilityType = iota
	VisibilityLocal     VisibilityType = iota
)

var VisibilityTypeStrs = map[VisibilityType]string{
	VisibilityDefault:   "<package-public>",
	VisibilityPrivate:   "private",
	VisibilityPublic:    "public",
	VisibilityProtected: "protected",
	VisibilityLocal:     "<local>",
}

type JavaTypeType int

const (
	JavaTypePrimitive  JavaTypeType = iota
	JavaTypeClass      JavaTypeType = iota
	JavaTypeInterface  JavaTypeType = iota
	JavaTypeEnum       JavaTypeType = iota
	JavaTypeRecord     JavaTypeType = iota
	JavaTypeAnnotation JavaTypeType = iota

	// JavaTypeLSPClass is a special kind of Java type used internally by the LSP to represent classes
	JavaTypeLSPClass JavaTypeType = iota
	// JavaTypeLSPConstructor is a special kind of Java type used internally by the LSP to represent constructors
	JavaTypeLSPConstructor JavaTypeType = iota
	// JavaTypeLSPMethod is a special kind of Java type used internally by the LSP to represent methods
	JavaTypeLSPMethod JavaTypeType = iota
	// JavaTypeLSPAny is a special Java type used internally by the LSP to represent error states, which shouldn't create further type errors
	JavaTypeLSPAny JavaTypeType = iota
)

const (
	TypeNameLSPClass       = "__LSPClass__"
	TypeNameLSPConstructor = "__LSPConstructor__"
	TypeNameLSPMethod      = "__LSPMethod__"
	TypeNameLSPAny         = "__LSPAny__"
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
	Fields       []*JavaField
	Methods      []*JavaMethod

	// TODO add generic constraints

	// For a *concrete* generic type (e.g. List<String>), this contains the type arguments (String).
	// For an *abstract* generic type (e.g. List<T>), this is nil.
	// NOTE: These are only used for method resolution currently
	GenericArgs []*JavaType

	// Definition stores where this symbol is defined in the code.
	// Is nil for built-in/library types.
	Definition *loc.CodeLocation
	// Usages stores all code locations where this type is referenced.
	Usages []loc.CodeLocation

	Visibility VisibilityType
	Type       JavaTypeType
}

func NewJavaType(name string, ppackage string, visibility VisibilityType, ttype JavaTypeType, definition *loc.CodeLocation) *JavaType {
	return &JavaType{
		Name:         name,
		Package:      ppackage,
		Module:       "",
		Extends:      make([]*JavaType, 0),
		Implements:   make([]*JavaType, 0),
		Constructors: make([]*JavaConstructor, 0),
		Fields:       make([]*JavaField, 0),
		Methods:      make([]*JavaMethod, 0),
		GenericArgs:  nil,
		Definition:   definition,
		Usages:       make([]loc.CodeLocation, 0),
		Visibility:   visibility,
		Type:         ttype,
	}
}

func NewPrimitiveType(name string) *JavaType {
	return NewJavaType(name, "", VisibilityPublic, JavaTypePrimitive, nil)
}

// Compile-time check that JavaType implements JavaSymbol interface
var _ JavaSymbol = (*JavaType)(nil)

func (jt *JavaType) Kind() JavaSymbolKind {
	return JavaSymbolType
}

func (jt *JavaType) PackageName() string {
	return jt.Package
}

func (jt *JavaType) ShortName() string {
	return jt.Name
}

func (jt *JavaType) FullName() string {
	genericsStr := ""
	if jt.GenericArgs != nil && len(jt.GenericArgs) > 0 {
		genericsStr = strings.Join(util.Map(jt.GenericArgs, func(t *JavaType) string {
			return t.FullName()
		}), ",")

		genericsStr = "<" + genericsStr + ">"
	}

	// TODO use this when we can handle package names
	//var mainName string
	//if jt.Package != "" {
	//	mainName = fmt.Sprintf("%s.%s", jt.Package, jt.Name)
	//} else {
	//	mainName =
	//}

	return jt.Name + genericsStr
}

func (jt *JavaType) GetVisibility() VisibilityType {
	return jt.Visibility
}

func (jt *JavaType) GetDefinition() *loc.CodeLocation {
	return jt.Definition
}

func (jt *JavaType) GetUsages() []loc.CodeLocation {
	return jt.Usages
}

func (jt *JavaType) AddUsage(location loc.CodeLocation) {
	jt.Usages = append(jt.Usages, location)
}

func (jt *JavaType) GetType() *JavaType {
	// Create a new type just for this class
	t := NewJavaType(TypeNameLSPClass, "", VisibilityPublic, JavaTypeLSPClass, nil)
	t.GenericArgs = []*JavaType{jt}
	return t
	// Maybe TODO: If this type already exists, use the existing one, otherwise use the newly created one
}

func (jt *JavaType) GetClassName() string {
	if jt.Type == JavaTypeLSPClass {
		referringType := jt.GenericArgs[0]
		return referringType.Name
	}
	return jt.Name
}

func (jt *JavaType) LookupMember(name string) JavaSymbol {
	if jt.Type == JavaTypeLSPClass {
		return jt.lookupStaticMember(name)
	}

	// First check fields
	idx := slices.IndexFunc(jt.Fields, func(field *JavaField) bool {
		return field.Name == name
	})
	if idx != -1 {
		return jt.Fields[idx]
	}

	// Then check methods
	// TODO handle overrides
	idx = slices.IndexFunc(jt.Methods, func(method *JavaMethod) bool {
		return method.Name == name
	})
	if idx != -1 {
		return jt.Methods[idx]
	}

	// Go to parent class/interfaces and see if any of them have the field
	for _, supertype := range jt.Extends {
		member := supertype.LookupMember(name)
		if member != nil {
			return member
		}
	}

	// Not found
	return nil
}

func (jt *JavaType) lookupStaticMember(name string) JavaSymbol {
	referringType := jt.GenericArgs[0]
	return referringType.LookupMember(name)
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
		supers = append(supers, e.AllSuperClasses()...)
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
	if jt.Type == JavaTypeLSPAny {
		// special "any" type coerces to any other type
		return true
	}

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
	ParentType *JavaType

	// Definition stores where this symbol is defined in the code.
	// Is nil for built-in/library types.
	Definition *loc.CodeLocation
	// Usages stores all code locations where this type is referenced.
	Usages []loc.CodeLocation

	Visibility VisibilityType
	IsStatic   bool
	IsFinal    bool
}

var _ JavaSymbol = (*JavaField)(nil)

func (jf *JavaField) Kind() JavaSymbolKind {
	return JavaSymbolField
}

func (jf *JavaField) PackageName() string {
	return jf.ParentType.Package
}

func (jf *JavaField) ShortName() string {
	return jf.Name
}

func (jf *JavaField) FullName() string {
	return fmt.Sprintf("%s.%s", jf.ParentType.FullName(), jf.Name)
}

func (jf *JavaField) GetVisibility() VisibilityType {
	return jf.Visibility
}

func (jf *JavaField) GetDefinition() *loc.CodeLocation {
	return jf.Definition
}

func (jf *JavaField) GetUsages() []loc.CodeLocation {
	return jf.Usages
}

func (jf *JavaField) AddUsage(location loc.CodeLocation) {
	jf.Usages = append(jf.Usages, location)
}

func (jf *JavaField) GetType() *JavaType {
	return jf.Type
}

func (jf *JavaField) String() string {
	return fmt.Sprintf("%s %s%s %s", VisibilityTypeStrs[jf.Visibility], getStaticStr(jf.IsStatic), jf.ParentType.Name, jf.Name)
}

type JavaConstructor struct {
	ParentType *JavaType
	Params     []*JavaParameter

	// Definition stores where this constructor is defined in the code.
	// Is nil for built-in/library types.
	Definition *loc.CodeLocation
	// Usages stores all code locations where this constructor is referenced.
	Usages []loc.CodeLocation

	Visibility VisibilityType
}

var _ JavaSymbol = (*JavaConstructor)(nil)

func (jc *JavaConstructor) Kind() JavaSymbolKind {
	return JavaSymbolConstructor
}

func (jc *JavaConstructor) PackageName() string {
	return jc.ParentType.Package
}

func (jc *JavaConstructor) ShortName() string {
	return fmt.Sprintf("%s(%s)", jc.ParentType.Name, strings.Join(util.MapToString(jc.Params), ","))
}

func (jc *JavaConstructor) FullName() string {
	return fmt.Sprintf("%s.%s", jc.ParentType.FullName(), jc.ShortName())
}

func (jc *JavaConstructor) GetVisibility() VisibilityType {
	return jc.Visibility
}

func (jc *JavaConstructor) GetDefinition() *loc.CodeLocation {
	return jc.Definition
}

func (jc *JavaConstructor) GetUsages() []loc.CodeLocation {
	return jc.Usages
}

func (jc *JavaConstructor) AddUsage(location loc.CodeLocation) {
	jc.Usages = append(jc.Usages, location)
}

func (jc *JavaConstructor) GetType() *JavaType {
	// Create a new type just for this constructor that reflects its receiver type, return type, and parameters

	t := NewJavaType(TypeNameLSPConstructor, "", VisibilityPublic, JavaTypeLSPConstructor, nil)

	t.GenericArgs = make([]*JavaType, 1+len(jc.Params))
	// Receiver type (also return type for constructor)
	t.GenericArgs[0] = jc.ParentType

	for i, param := range jc.Params {
		// TODO if it's a varargs param, change it into the array type
		t.GenericArgs[i+1] = param.Type
	}

	return t

	// Maybe TODO: If this type already exists, use the existing one, otherwise use the newly created one
}

type JavaMethod struct {
	Name       string
	ParentType *JavaType
	ReturnType *JavaType
	Params     []*JavaParameter

	// Definition stores where this method is defined in the code.
	// Is nil for built-in/library types.
	Definition *loc.CodeLocation
	// Usages stores all code locations where this method is referenced.
	Usages []loc.CodeLocation

	Visibility VisibilityType
	IsStatic   bool
}

var _ JavaSymbol = (*JavaMethod)(nil)

func (jm *JavaMethod) Kind() JavaSymbolKind {
	return JavaSymbolMethod
}

func (jm *JavaMethod) PackageName() string {
	return jm.ParentType.Package
}

func (jm *JavaMethod) ShortName() string {
	return jm.Name
}

func (jm *JavaMethod) NameWithArgs() string {
	return fmt.Sprintf("%s(%s)", jm.Name, strings.Join(util.MapToString(jm.Params), ","))
}

func (jm *JavaMethod) FullName() string {
	var returnTypeName string
	if jm.ReturnType == nil {
		returnTypeName = "void"
	} else {
		returnTypeName = jm.ReturnType.ShortName()
	}

	return fmt.Sprintf("%s %s.%s", returnTypeName, jm.ParentType.FullName(), jm.NameWithArgs())
}

func (jm *JavaMethod) GetVisibility() VisibilityType {
	return jm.Visibility
}

func (jm *JavaMethod) GetDefinition() *loc.CodeLocation {
	return jm.Definition
}

func (jm *JavaMethod) GetUsages() []loc.CodeLocation {
	return jm.Usages
}

func (jm *JavaMethod) AddUsage(location loc.CodeLocation) {
	jm.Usages = append(jm.Usages, location)
}

func (jm *JavaMethod) GetType() *JavaType {
	// Create a new type just for this method that reflects its receiver type, return type, and parameters

	t := NewJavaType(TypeNameLSPMethod, "", VisibilityPublic, JavaTypeLSPMethod, nil)

	t.GenericArgs = make([]*JavaType, 2+len(jm.Params))
	t.GenericArgs[0] = jm.ParentType
	t.GenericArgs[1] = jm.ReturnType
	for i, param := range jm.Params {
		// TODO if it's a varargs param, change it into the array type
		t.GenericArgs[i+2] = param.Type
	}

	return t

	// Maybe TODO: If this type already exists, use the existing one, otherwise use the newly created one
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

func (jp *JavaParameter) String() string {
	return fmt.Sprintf("%s %s", jp.Type.Name, jp.Name)
}

type JavaLocal struct {
	Name         string
	Type         *JavaType
	ParentMethod *JavaMethod

	// Definition stores where this method is defined in the code.
	Definition *loc.CodeLocation
	// Usages stores all code locations where this method is referenced.
	Usages []loc.CodeLocation
}

func NewJavaLocal(name string, ttype *JavaType, parentMethod *JavaMethod, definition loc.CodeLocation) *JavaLocal {
	return &JavaLocal{
		Name:         name,
		Type:         ttype,
		ParentMethod: parentMethod,
		Definition:   &definition,
		Usages:       make([]loc.CodeLocation, 0),
	}
}

var _ JavaSymbol = (*JavaLocal)(nil)

func (jl *JavaLocal) Kind() JavaSymbolKind {
	return JavaSymbolLocal
}

func (jl *JavaLocal) PackageName() string {
	return jl.ParentMethod.ParentType.Package
}

func (jl *JavaLocal) ShortName() string {
	return jl.Name
}

func (jl *JavaLocal) FullName() string {
	return fmt.Sprintf("%s %s (local var in %s)", jl.Type.FullName(), jl.Name, jl.ParentMethod.FullName())
}

func (jl *JavaLocal) GetVisibility() VisibilityType {
	return VisibilityLocal
}

func (jl *JavaLocal) GetDefinition() *loc.CodeLocation {
	return jl.Definition
}

func (jl *JavaLocal) GetUsages() []loc.CodeLocation {
	return jl.Usages
}

func (jl *JavaLocal) AddUsage(location loc.CodeLocation) {
	jl.Usages = append(jl.Usages, location)
}

func (jl *JavaLocal) GetType() *JavaType {
	return jl.Type
}
