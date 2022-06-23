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

func getStaticStr(isStatic bool) string {
	if isStatic {
		return "static "
	} else {
		return ""
	}
}

type JavaType struct {
	Name        string
	PackageName string
	Visibility  VisibilityType
	Fields      []*JavaField
	Methods     []*JavaMethod
}

func (jt *JavaType) String() string {
	return jt.Name
}

func getOrCreateBuiltinType(name string) *JavaType {
	jtype, ok := BuiltinTypes[name]
	if !ok {
		jtype = &JavaType{
			Name:       name,
			Visibility: VisibilityPublic,
		}
		BuiltinTypes[name] = jtype
	}

	return jtype
}

type JavaField struct {
	Name       string
	Visibility VisibilityType
	Type       *JavaType
	IsStatic   bool
	IsFinal    bool
}

func (jf *JavaField) String() string {
	return fmt.Sprintf("%s %s%s %s", VisibilityTypeStrs[jf.Visibility], getStaticStr(jf.IsStatic), jf.Type.Name, jf.Name)
}

type JavaMethod struct {
	Name       string
	Visibility VisibilityType
	ReturnType *JavaType
	Arguments  []*JavaArgument
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
	Name string
	Type *JavaType
}

func (ja *JavaArgument) String() string {
	return fmt.Sprintf("%s %s", ja.Type.Name, ja.Name)
}
