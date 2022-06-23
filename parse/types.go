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

type JavaField struct {
	Name       string
	Visibility VisibilityType
}

type JavaMethod struct {
	Name       string
	Visibility VisibilityType
	ReturnType *JavaType
	Arguments  []*JavaType
	IsStatic   bool
}

func (jm *JavaMethod) String() string {
	argStr := ""
	if jm.Arguments != nil {
		argStr = strings.Join(util.MapToString(jm.Arguments), ", ")
	}

	staticStr := ""
	if jm.IsStatic {
		staticStr = "static "
	}

	return fmt.Sprintf("%s %s%s %s(%s)", VisibilityTypeStrs[jm.Visibility], staticStr, jm.ReturnType, jm.Name, argStr)
}
