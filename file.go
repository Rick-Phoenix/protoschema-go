package protoschema

import (
	"errors"
	"fmt"
	"path"

	u "github.com/Rick-Phoenix/goutils"
)

// Function that receives the file data after processing the its schema. If it returns an error, this will be marked as fatal at the very last moment, in order to accumulate all the errors in the schemas and report them.
type FileHook func(d FileData) error

// The struct for the File schema. It should be created with the constructor from the ProtoPackage instance except for defining file data for imported message types that have not been created from this library.
type FileSchema struct {
	// Automatically set when created with the constructor from a ProtoPackage instance.
	Package *ProtoPackage
	// The name of the file. The ".proto" suffix will be added automatically by the constructor and the getter.
	Name string
	// Imports required by the components of the file will be added automatically. This can be used to add extra imports if necessary.
	Imports Set
	// The protobuf extensions for this file.
	Extensions Extensions
	// Top level options.
	Options  []ProtoOption
	enums    []*EnumGroup
	messages []*MessageSchema
	services []*ServiceSchema
	// Overrides the package-level FileHook, if defined.
	Hook FileHook
	// Custom map to store data that can be used with hooks.
	Metadata map[string]any
}

// The struct containing all the results from processing the components of the file. Will be passed to the FileHook after being generated if it's defined.
type FileData struct {
	Package    *ProtoPackage
	Name       string
	Imports    Set
	Extensions Extensions
	Options    []ProtoOption
	Enums      []EnumGroup
	Messages   []MessageData
	Services   []ServiceData
	Metadata   map[string]any
}

// Accesses the name safely, and adds the ".proto" suffix if missing.
func (f *FileSchema) GetName() string {
	if f == nil {
		return ""
	}

	if f.Name == "" {
		return ""
	}

	return addMissingSuffix(f.Name, ".proto")
}

// Returns the base path of its package, joined with the file name.
func (f *FileSchema) GetImportPath() string {
	if f == nil {
		return ""
	}

	name := f.GetName()

	if name == "" {
		return ""
	}

	return path.Join(f.Package.GetBasePath(), name)
}

// Creates a new MessageSchema instance, automatically setting its Package, File and ImportPath fields.
// If a Hook is missing in the schema, it assigns the global MessageHook (if defined).
func (f *FileSchema) NewMessage(s MessageSchema) *MessageSchema {
	s.Package = f.Package
	s.File = f
	s.ImportPath = path.Join(f.Package.BasePath, f.Name)
	f.messages = append(f.messages, &s)
	if s.Hook == nil {
		s.Hook = s.Package.messageHook
	}
	if s.ConverterFunc == nil && f.Package != nil {
		s.ConverterFunc = f.Package.converterFunc
	}
	return &s
}

// Creates a new ServiceSchema instance, automatically setting its Package and File fields.
// If a Hook is missing in the schema, it assigns the global ServiceHook (if defined).
func (f *FileSchema) NewService(s ServiceSchema) *ServiceSchema {
	s.Package = f.Package
	s.File = f
	f.services = append(f.services, &s)
	if s.Hook == nil && f.Package != nil {
		s.Hook = f.Package.serviceHook
	}
	return &s
}

// Creates a new EnumGroup instance, automatically setting its Package and File fields.
func (f *FileSchema) NewEnum(e EnumGroup) *EnumGroup {
	e.File = f
	e.Package = f.Package
	f.enums = append(f.enums, &e)
	return &e
}

func (f *FileSchema) build() (FileData, error) {
	imports := make(Set)

	file := FileData{
		Package:    f.Package,
		Imports:    imports,
		Extensions: f.Extensions,
		Options:    f.Options,
		Name:       f.Name,
		Metadata:   f.Metadata,
		Enums:      u.ToValSlice(f.enums),
	}

	if len(f.Extensions.File)+len(f.Extensions.Service)+len(f.Extensions.Message)+len(f.Extensions.Field)+len(f.Extensions.OneOf) > 0 {
		imports["google/protobuf/descriptor.proto"] = present
	}

	var messageErrors error

	for _, m := range f.messages {
		var errAgg error

		message, err := m.build(imports)
		errAgg = errors.Join(errAgg, err)
		file.Messages = append(file.Messages, message)

		if errAgg != nil {
			messageErrors = errors.Join(messageErrors, indentErrors(fmt.Sprintf("Errors for the %q message schema", m.GetName()), errAgg))
		}
	}

	for _, serv := range f.services {
		file.Services = append(file.Services, serv.build(imports))
	}

	if f.Hook != nil {
		err := f.Hook(file)
		if err != nil {
			fmt.Printf("Error in file hook:\n")
			return file, err
		}
	}

	return file, messageErrors
}
