package typ

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"java-mini-ls-go/parse/loc"
	"java-mini-ls-go/util"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

type TypeMap struct {
	contents map[string]*JavaType
}

func NewTypeMap() *TypeMap {
	return &TypeMap{contents: make(map[string]*JavaType)}
}

func (tm *TypeMap) Add(t *JavaType) {
	tm.contents[t.FullName()] = t
}

func (tm *TypeMap) Get(s string) *JavaType {
	return tm.contents[s]
}

func (tm *TypeMap) Size() int {
	return len(tm.contents)
}

// The actual built-in types map that we load into
var builtinTypes *TypeMap

// The mutex is probably not necessary, but good to have just in case
var builtinTypesMutex sync.Mutex

func LoadBuiltinTypes() (*TypeMap, error) {
	builtinTypesMutex.Lock()
	defer builtinTypesMutex.Unlock()

	if builtinTypes != nil {
		return builtinTypes, nil
	}

	fmt.Println("Loading Java standard library types ...")
	now := time.Now()
	startTime := now.UnixMilli()

	err := loadBuiltinTypesFromDisk()
	if err != nil {
		err = errors.Wrap(err, "Error loading Java standard library types")
		fmt.Println(err)
		return nil, err
	}

	now = time.Now()
	endTime := now.UnixMilli()
	duration := endTime - startTime
	fmt.Printf("Loaded %d Java standard library types in %dms\n", builtinTypes.Size(), duration)

	return builtinTypes, nil
}

// JSON-based types for loading from file

type javaJsonType struct {
	Name         string                `json:"name"`
	Type         string                `json:"type"`
	Module       string                `json:"module"`
	Package      string                `json:"package"`
	Extends      []string              `json:"extends"`
	Implements   []string              `json:"implements"`
	Fields       []javaJsonField       `json:"fields"`
	Methods      []javaJsonMethod      `json:"methods"`
	Constructors []javaJsonConstructor `json:"constructors"`

	// TODO add generics
}

func (jjt *javaJsonType) FullName() string {
	// Should mimic the output of JavaType.FullName()
	return jjt.Name
}

type javaJsonField struct {
	Name        string   `json:"name"`
	Modifiers   []string `json:"modifiers"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
}

type javaJsonMethod struct {
	Name        string        `json:"name"`
	Modifiers   []string      `json:"modifiers"`
	Type        string        `json:"type"`
	Description string        `json:"description"`
	Args        []javaJsonArg `json:"args"`
}

type javaJsonArg struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type javaJsonConstructor struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Args        []javaJsonArg `json:"args"`
}

func loadBuiltinTypesFromDisk() error {
	jsonTypes, err := readJsonFromDisk()
	if err != nil {
		return errors.Wrapf(err, "Error reading JSON from disk")
	}

	builtinTypes = NewTypeMap()

	// add primitive types before we load the rest of the types
	AddPrimitiveTypes(builtinTypes)

	err = loadJsonTypes(jsonTypes)
	if err != nil {
		return errors.Wrapf(err, "Error parsing JSON")
	}

	return nil
}

func readJsonFromDisk() ([]javaJsonType, error) {
	stdlibJsonPath, err := getStdlibJsonPath()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting path of Java stdlib json file")
	}

	filename := filepath.Join(stdlibJsonPath, "java_stdlib.json")
	jsonFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {
			fmt.Println("Error closing json file: ", err)
		}
	}(jsonFile)

	// Currently we read the entire file into memory. If it gets too big, we may want to stream it instead:
	// https://stackoverflow.com/questions/31794355/decode-large-stream-json
	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	var types []javaJsonType
	err = json.Unmarshal(bytes, &types)
	if err != nil {
		return nil, err
	}

	return types, nil
}

func getStdlibJsonPath() (string, error) {
	// Determine filename from caller location
	_, filename, _, _ := runtime.Caller(0)
	idx := strings.LastIndex(filename, "java-mini-ls-go")
	if idx != -1 {
		filename = filename[0:idx+len("java-mini-ls-go")] + "/"
	}

	absPath, err := filepath.Abs(filepath.Dir(filename))
	if err != nil {
		return "", err
	}

	return absPath, nil
}

func AddPrimitiveTypes(typeMap *TypeMap) {
	primitives := []string{
		"byte",
		"short",
		"int",
		"long",
		"float",
		"double",
		"boolean",
		"char",
	}

	for _, name := range primitives {
		typeMap.Add(NewPrimitiveType(name))
	}
}

func convertJsonTypeType(jsonTypeType string) JavaTypeType {
	switch jsonTypeType {
	case "class":
		return JavaTypeClass
	case "interface":
		return JavaTypeInterface
	case "enum":
		return JavaTypeEnum
	case "annotation":
		return JavaTypeAnnotation
	case "record":
		return JavaTypeRecord
	default:
		fmt.Println("Unknown Java JSON type type: ", jsonTypeType)
		return JavaTypeClass
	}
}

func convertJsonField(parentType *JavaType, jsonField javaJsonField) *JavaField {
	return &JavaField{
		Name:       jsonField.Name,
		ParentType: parentType,
		Type:       getOrCreateBuiltinType(jsonField.Type),
		Visibility: VisibilityPublic,
		IsStatic:   slices.Contains(jsonField.Modifiers, "static"),
		IsFinal:    slices.Contains(jsonField.Modifiers, "final"),
		Definition: nil,
		Usages:     []loc.CodeLocation{},
	}
}

func convertJsonMethod(parentType *JavaType, jsonMethod javaJsonMethod) *JavaMethod {
	return &JavaMethod{
		Name:       jsonMethod.Name,
		ParentType: parentType,
		ReturnType: getOrCreateBuiltinType(jsonMethod.Type),
		Params:     util.Map(jsonMethod.Args, toArg),
		Visibility: VisibilityPublic,
		IsStatic:   slices.Contains(jsonMethod.Modifiers, "static"),
		Definition: nil,
		Usages:     []loc.CodeLocation{},
	}
}

// Loads provided JSON types into builtinTypes map
func loadJsonTypes(jsonTypes []javaJsonType) error {
	// First, get just the bare types defined
	for _, jsonType := range jsonTypes {
		newType := NewJavaType(jsonType.Name, jsonType.Package, VisibilityPublic, convertJsonTypeType(jsonType.Type), nil)
		builtinTypes.Add(newType)
	}

	// Next, fill in extends/implements references
	for _, jsonType := range jsonTypes {
		ttype := builtinTypes.Get(jsonType.Name)
		if jsonType.Extends != nil {
			ttype.Extends = util.Map(jsonType.Extends, getOrCreateBuiltinType)
		}
		if jsonType.Implements != nil {
			ttype.Implements = util.Map(jsonType.Implements, getOrCreateBuiltinType)
		}
	}

	// Next, fill in the constructors
	for _, jsonType := range jsonTypes {
		constructors := make([]*JavaConstructor, 0, len(jsonType.Constructors))

		parentType := builtinTypes.Get(jsonType.FullName())

		for _, jsonConstructor := range jsonType.Constructors {
			constructors = append(constructors, &JavaConstructor{
				ParentType: parentType,
				Params:     util.Map(jsonConstructor.Args, toArg),
				Definition: nil,
				Usages:     []loc.CodeLocation{},
				Visibility: VisibilityPublic,
			})
		}

		parentType.Constructors = constructors
	}

	// Next, fill in the fields & methods
	for _, jsonType := range jsonTypes {
		parentType := builtinTypes.Get(jsonType.FullName())
		parentType.Fields = util.Map(jsonType.Fields, func(jsonField javaJsonField) *JavaField {
			return convertJsonField(parentType, jsonField)
		})
		parentType.Methods = util.Map(jsonType.Methods, func(jsonMethod javaJsonMethod) *JavaMethod {
			return convertJsonMethod(parentType, jsonMethod)
		})
	}

	return nil
}

func toArg(arg javaJsonArg) *JavaParameter {
	return &JavaParameter{
		Name:      arg.Name,
		Type:      getOrCreateBuiltinType(arg.Type),
		IsVarargs: false,
	}
}

// Function for getting/creating builtin types that *should* exist but for some reason we haven't parsed them.
// Creates a basic placeholder type for them.
func getOrCreateBuiltinType(name string) *JavaType {
	jtype := builtinTypes.Get(name)
	if jtype == nil {
		fmt.Println("Creating built-in type: ", name)
		jtype = NewJavaType(name, "", VisibilityPublic, JavaTypeClass, nil)
		builtinTypes.Add(jtype)
	}

	return jtype
}
