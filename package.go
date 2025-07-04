package protoschema

import (
	"errors"
	"fmt"
	"log"
	"maps"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
)

// The configuration for a ProtoPackage instance
type ProtoPackageConfig struct {
	// The name of the package, i.e. "myapp.v1".
	Name string
	// The path to the root of the proto project.
	ProtoRoot string
	// The full path to the package of the generated go files.
	GoPackage string
	// The go module for this project.
	GoModule string
	// (Default: "gen/converter") The output directory for the converter file. The last part will be the name of the converter package.
	ConverterOutputDir string
	// Function that runs after each file schema is processed (can be overridden at the schema level).
	FileHook FileHook
	// Function that runs after each service schema is processed (can be overridden at the schema level).
	ServiceHook ServiceHook
	// Function that runs after each message schema is processed (can be overridden at the schema level).
	MessageHook MessageHook
	// Function that runs after each oneof schema is processed (can be overridden at the schema level).
	OneofHook OneofHook
	// If undefined, the package will use a default function to try and automatically generate a file that contains functions to automatically convert a struct from the message's Model type (for example, a struct coming from a database or from another api) to the generated type for that message.
	// If defined, this function will receive a rich set of data for each message field to define its own logic for generating files or performing custom actions.
	// It can also be overridden for a single message.
	ConverterFunc ConverterFunc
}

// The ProtoPackage struct, which holds the data and methods for file generation (if created with the constructor). Can also be used without the constructor to define the package data for imported message types.
type ProtoPackage struct {
	// The name of the package, i.e. "myapp.v1".
	Name string
	// If accessed with the getter (or the ProtoPackage instance was created with the constructor), it defaults to the package name with slashes instead of dots, as per the proto convention ("myapp.v1" -> "myapp/v1")
	BasePath string
	// The full path to the package of the generated go files.
	GoPackagePath string
	// If accessed with the getter (or the ProtoPackage instance was created with the constructor), it defaults to the last part of the go package path.
	GoPackageName      string
	protoRoot          string
	goModule           string
	converterOutputDir string
	converterPackage   string
	protoOutputDir     string
	fileHook           FileHook
	serviceHook        ServiceHook
	messageHook        MessageHook
	oneofHook          OneofHook
	tmpl               *template.Template
	fileSchemas        []*FileSchema
	converter          converterData
	converterFunc      ConverterFunc
}

type StoreData struct {
	PkgPath string
	Methods map[string]StoreMethod
}

type StoreMethod struct {
	Name       string
	ReturnType string
}

// Returns the name of the package, defaulting to an empty string if the pointer is nil.
func (hb *ProtoPackage) GetName() string {
	if hb == nil {
		return ""
	}

	return hb.Name
}

// Accesses the go package name and uses the default value if it's missing.
func (hb *ProtoPackage) GetGoPackageName() string {
	if hb == nil {
		return ""
	}

	if hb.GoPackageName == "" {
		if goPkgBase := path.Base(hb.GoPackagePath); goPkgBase != "." {
			return goPkgBase
		}
	}

	return hb.GoPackageName
}

// Accesses the base path of the proto package, inferring it (assuming the conventional structure "myapp.v1" -> "myapp/v1") from the package's name if not explicitely defined.
func (hb *ProtoPackage) GetBasePath() string {
	if hb == nil {
		return ""
	}

	if hb.BasePath == "" {
		pkgPath := strings.ReplaceAll(hb.Name, ".", "/")
		hb.BasePath = pkgPath
		return pkgPath
	}

	return hb.BasePath
}

// Returns the full path to the package of the generated go files.
func (hb *ProtoPackage) GetGoPackagePath() string {
	if hb == nil {
		return ""
	}

	return hb.GoPackagePath
}

func extractMethods(model reflect.Type) map[string]StoreMethod {
	output := make(map[string]StoreMethod)

	for i := range model.NumMethod() {
		method := model.Method(i)
		data := StoreMethod{}
		data.Name = method.Name
		if method.Type.NumOut() > 0 {
			outType := method.Type.Out(0)
			if outType.Kind() == reflect.Pointer {
				outType = outType.Elem()
			}
			data.ReturnType = outType.Name()
		}
		output[data.Name] = data
	}

	return output
}

// The constructor for a ProtoPackage instance.
func NewProtoPackage(conf ProtoPackageConfig) *ProtoPackage {
	p := &ProtoPackage{
		Name:               conf.Name,
		BasePath:           strings.ReplaceAll(conf.Name, ".", "/"),
		GoPackagePath:      conf.GoPackage,
		protoRoot:          conf.ProtoRoot,
		goModule:           conf.GoModule,
		converterOutputDir: conf.ConverterOutputDir,
		fileHook:           conf.FileHook,
		serviceHook:        conf.ServiceHook,
		messageHook:        conf.MessageHook,
		oneofHook:          conf.OneofHook,
		converterFunc:      conf.ConverterFunc,
	}

	if conf.Name == "" {
		log.Fatalf("Missing proto package definition.")
	}

	if conf.GoPackage == "" {
		log.Fatalf("Missing go package definition.")
	}

	p.GoPackageName = path.Base(conf.GoPackage)

	p.protoOutputDir = filepath.Join(conf.ProtoRoot, p.BasePath)

	if p.converterOutputDir == "" {
		p.converterOutputDir = "gen/converter"
	}

	tmpl, err := template.New("protoTemplates").Funcs(funcMap).ParseFS(templateFS, "templates/*")
	if err != nil {
		fmt.Print(fmt.Errorf("Failed to initiate tmpl instance for the generator: %w", err))
		os.Exit(1)
	}
	p.tmpl = tmpl

	p.converterPackage = filepath.Base(p.converterOutputDir)
	converter := converterData{
		Package:   p.converterPackage,
		GoPackage: p.GoPackageName,
		Imports:   Set{p.GoPackagePath: present}, RepeatedConverters: make(Set),
	}
	p.converter = converter

	return p
}

// Adds a file to the package and returns a pointer to it.
func (hb *ProtoPackage) NewFile(s FileSchema) *FileSchema {
	newFile := &FileSchema{
		Name:       s.Name + ".proto",
		Package:    hb,
		Imports:    make(Set),
		Options:    s.Options,
		Extensions: s.Extensions,
		enums:      s.enums,
		messages:   s.messages,
		services:   s.services,
		Hook:       s.Hook,
	}
	maps.Copy(newFile.Imports, s.Imports)
	if s.Hook == nil {
		s.Hook = hb.fileHook
	}
	hb.fileSchemas = append(hb.fileSchemas, newFile)
	return newFile
}

// Processes all the files' data and returns it. This is called automatically when .Generate() is called.
// In most cases it's better to use the FileHook to perform custom actions on the data, but this can also be used to collect all the processed data and use it directly.
func (hb *ProtoPackage) BuildFiles() []FileData {
	out := []FileData{}
	var fileErrors error

	for _, f := range hb.fileSchemas {
		file, err := f.build()
		if err != nil {
			fileErrors = errors.Join(fileErrors, indentErrors(fmt.Sprintf("Errors in the file %q", f.Name), err))
		}

		out = append(out, file)

	}

	if fileErrors != nil {
		fmt.Printf("  ❌ The following errors occurred:\n")
		fmt.Print(fileErrors.Error())
		os.Exit(1)
	}

	return out
}
