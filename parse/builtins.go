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

// The actual built-in types map that we load into
var builtinTypes map[string]*JavaType

// The mutex is probably not necessary, but good to have just in case
var builtinTypesMutex sync.Mutex

func LoadBuiltinTypes() (map[string]*JavaType, error) {
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
	fmt.Printf("Done loading Java standard library types (took %dms)\n", duration)

	return builtinTypes, nil
}

// JSON-based types for loading from file

type javaJsonType struct {
	Name         string                `json:"name"`
	Module       string                `json:"module"`
	Package      string                `json:"package"`
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

	builtinTypes = make(map[string]*JavaType)

	err = loadJsonTypes(jsonTypes)
	if err != nil {
		return errors.Wrapf(err, "Error parsing JSON")
	}

	return nil
}

func readJsonFromDisk() ([]javaJsonType, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	absPath, err := filepath.Abs(filepath.Dir(execPath))
	if err != nil {
		return nil, err
	}

	filename := filepath.Join(absPath, "java_stdlib.json")
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
				Arguments:  util.Map(jsonMethod.Args, toArg),
				IsStatic:   slices.Contains(jsonMethod.Modifiers, "static"),
			}
		}

		builtinTypes[jsonType.Name].Methods = methods
	}

	return nil
}

func toArg(arg javaJsonArg) *JavaArgument {
	return &JavaArgument{
		Name: arg.Name,
		Type: getOrCreateBuiltinType(arg.Type),
	}
}

// Function for getting/creating builtin types that *should* exist but for some reason we haven't parsed them.
// Creates a basic placeholder type for them.
func getOrCreateBuiltinType(name string) *JavaType {
	jtype, ok := builtinTypes[name]
	if !ok {
		jtype = &JavaType{
			Name:       name,
			Visibility: VisibilityPublic,
		}
		builtinTypes[name] = jtype
	}

	return jtype
}
