package parse

import (
	"fmt"
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
	JavaTypeClass:     "class",
	JavaTypeInterface: "interface",
	JavaTypeEnum:      "enum",
	JavaTypeRecord:    "record",
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
