package parse

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
	"io/ioutil"
	"java-mini-ls-go/util"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type TypeMap map[string]*JavaType

// The actual built-in types map that we load into
var builtinTypes TypeMap

// The mutex is probably not necessary, but good to have just in case
var builtinTypesMutex sync.Mutex

func LoadBuiltinTypes() (TypeMap, error) {
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
	fmt.Printf("Loaded %d Java standard library types in %dms\n", len(builtinTypes), duration)

	return builtinTypes, nil
}

// JSON-based types for loading from file

type javaJsonType struct {
	Name         string                `json:"name"`
	Module       string                `json:"module"`
	Package      string                `json:"package"`
	Extends      []string              `json:"extends"`
	Implements   []string              `json:"implements"`
	Fields       []javaJsonField       `json:"fields"`
	Methods      []javaJsonMethod      `json:"methods"`
	Constructors []javaJsonConstructor `json:"constructors"`
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

	builtinTypes = make(TypeMap)

	// add primitive types before we load the rest of the types
	addPrimitiveTypes(builtinTypes)

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
	envPath := os.Getenv("JAVA_MINI_LS_STDLIB_PATH")
	if envPath != "" {
		return envPath, nil
	}

	// Env var not set, look up from executable location
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}

	absPath, err := filepath.Abs(filepath.Dir(execPath))
	if err != nil {
		return "", err
	}

	return absPath, nil
}

func addPrimitiveTypes(typeMap TypeMap) {
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
		typeMap[name] = &JavaType{
			Name:       name,
			Visibility: VisibilityPublic,
			Type:       JavaTypePrimitive,
		}
	}
}

// Loads provided JSON types into builtinTypes map
func loadJsonTypes(jsonTypes []javaJsonType) error {
	// First, get just the bare types defined
	for _, jsonType := range jsonTypes {
		builtinTypes[jsonType.Name] = &JavaType{
			Name:       jsonType.Name,
			Package:    jsonType.Package,
			Module:     jsonType.Module,
			Visibility: VisibilityPublic,
			Fields:     nil,
			Methods:    nil,
		}
	}

	// Next, fill in extends/implements references
	for _, jsonType := range jsonTypes {
		ttype := builtinTypes[jsonType.Name]
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

		for _, jsonConstructor := range jsonType.Constructors {
			constructors = append(constructors, &JavaConstructor{
				Visibility: VisibilityPublic,
				Arguments:  util.Map(jsonConstructor.Args, toArg),
			})
		}

		builtinTypes[jsonType.Name].Constructors = constructors
	}

	// Next, fill in the fields
	for _, jsonType := range jsonTypes {
		fields := make(map[string]*JavaField, len(jsonType.Fields))

		for _, jsonField := range jsonType.Fields {
			fields[jsonField.Name] = &JavaField{
				Name:       jsonField.Name,
				Visibility: VisibilityPublic,
				Type:       getOrCreateBuiltinType(jsonField.Type),
				IsStatic:   slices.Contains(jsonField.Modifiers, "static"),
				IsFinal:    slices.Contains(jsonField.Modifiers, "final"),
			}
		}

		builtinTypes[jsonType.Name].Fields = fields
	}

	// Next, fill in the methods
	for _, jsonType := range jsonTypes {
		methods := make(map[string]*JavaMethod, len(jsonType.Methods))

		for _, jsonMethod := range jsonType.Methods {
			methods[jsonMethod.Name] = &JavaMethod{
				Name:       jsonMethod.Name,
				Visibility: VisibilityPublic,
				ReturnType: getOrCreateBuiltinType(jsonMethod.Type),
				Params:     util.Map(jsonMethod.Args, toArg),
				IsStatic:   slices.Contains(jsonMethod.Modifiers, "static"),
			}
		}

		builtinTypes[jsonType.Name].Methods = methods
	}

	return nil
}

func toArg(arg javaJsonArg) *JavaParameter {
	return &JavaParameter{
		Name: arg.Name,
		Type: getOrCreateBuiltinType(arg.Type),
	}
}

// Function for getting/creating builtin types that *should* exist but for some reason we haven't parsed them.
// Creates a basic placeholder type for them.
func getOrCreateBuiltinType(name string) *JavaType {
	jtype, ok := builtinTypes[name]
	if !ok {
		//fmt.Println("Creating built-in type: ", name)
		jtype = &JavaType{
			Name:       name,
			Visibility: VisibilityPublic,
		}
		builtinTypes[name] = jtype
	}

	return jtype
}
