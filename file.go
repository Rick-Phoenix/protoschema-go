package schemabuilder

import (
	"errors"
	"fmt"
	"os"

	"github.com/labstack/gommon/log"
)

type FileSchema struct {
	PackageName string
	FileName    string
	Imports     Set
	Extensions  Extensions
	Options     []ProtoOption
	Enums       []EnumGroup
	Messages    map[string]MessageSchema
	Services    []ServiceSchema
}

type FileData struct {
	PackageName string
	FileName    string
	Imports     Set
	Extensions  Extensions
	Options     []ProtoOption
	Enums       []EnumGroup
	Messages    []MessageData
	Services    []ServiceData
}

func (f FileSchema) GetMessage(name string) MessageSchema {
	msg, found := f.Messages[name]
	if !found {
		log.Printf("Could not find message %q in file schema %q", name, f.FileName)
	}
	msg.ImportPath = f.FileName
	return msg
}

func (p *ProtoPackage) AddFile(s FileSchema) {
	p.files = append(p.files, s)
}

func (p *ProtoPackage) BuildFiles() []FileData {
	out := make([]FileData, len(p.files))
	var fileErrors error

	for _, s := range p.files {
		imports := make(Set)

		file := FileData{
			PackageName: s.PackageName,
			FileName:    s.PackageName,
			Imports:     imports,
			Extensions:  s.Extensions,
			Options:     s.Options,
			Enums:       s.Enums,
		}

		if len(s.Extensions.File)+len(s.Extensions.Service)+len(s.Extensions.Message)+len(s.Extensions.Field)+len(s.Extensions.OneOf) > 0 {
			imports["google/protobuf/descriptor.proto"] = present
		}

		var messageErrors error

		for _, m := range s.Messages {
			var errAgg error

			message, err := NewProtoMessage(m, imports)
			errAgg = errors.Join(errAgg, err)
			file.Messages = append(file.Messages, message)
			if message.Converter != nil {
				for _, v := range message.Converter.InternalRepeated {
					p.converters.RepeatedConverters[v] = present
				}
				for _, v := range message.Converter.Imports {
					p.converters.Imports[v] = present
				}

				p.converters.Converters = append(p.converters.Converters, *message.Converter)
			}

			if errAgg != nil {
				messageErrors = errors.Join(messageErrors, indentErrors(fmt.Sprintf("Errors for the %s message schema", m.Name), errAgg))
			}
		}

		if messageErrors != nil {
			fileErrors = errors.Join(fileErrors, indentErrors(fmt.Sprintf("Errors in the file %s", s.FileName), messageErrors))
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
