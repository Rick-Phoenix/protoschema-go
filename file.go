package schemabuilder

import (
	"errors"
	"fmt"
	"path"

	u "github.com/Rick-Phoenix/goutils"
)

type FileSchema struct {
	Package    *ProtoPackage
	Name       string
	Imports    Set
	Extensions Extensions
	Options    []ProtoOption
	enums      []*EnumGroup
	messages   []*MessageSchema
	services   []*ServiceSchema
	Hook       FileHook
	Metadata   map[string]any
}

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

func (f *FileSchema) GetName() string {
	if f == nil {
		return ""
	}

	if f.Name == "" {
		return ""
	}

	return addMissingSuffix(f.Name, ".proto")
}

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

func (f *FileSchema) NewService(s ServiceSchema) *ServiceSchema {
	s.Package = f.Package
	s.File = f
	f.services = append(f.services, &s)
	if s.Hook == nil && f.Package != nil {
		s.Hook = f.Package.serviceHook
	}
	return &s
}

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
			messageErrors = errors.Join(messageErrors, indentErrors(fmt.Sprintf("Errors for the %s message schema", m.Name), errAgg))
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
