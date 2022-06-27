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
	Name         string
	Package      string
	Module       string
	Constructors []*JavaConstructor
	Fields       map[string]*JavaField
	Methods      map[string]*JavaMethod
	Visibility   VisibilityType
	Type         JavaTypeType
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
	Arguments  []*JavaArgument
}

type JavaMethod struct {
	Name       string
	ReturnType *JavaType
	Arguments  []*JavaArgument
	Visibility VisibilityType
	IsStatic   bool
}

func (jm *JavaMethod) String() string {
	argStr := ""
	if jm.Arguments != nil {
		argStr = strings.Join(util.MapToString(jm.Arguments), ", ")
	}

	return fmt.Sprintf("%s %s%s %s(%s)", VisibilityTypeStrs[jm.Visibility], getStaticStr(jm.IsStatic), jm.ReturnType, jm.Name, argStr)
}

type JavaArgument struct {
	Name      string
	Type      *JavaType
	IsVarargs bool
}

func (ja *JavaArgument) String() string {
	return fmt.Sprintf("%s %s", ja.Type.Name, ja.Name)
}
