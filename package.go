package schemabuilder

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

type ProtoPackageConfig struct {
	Name               string
	ProtoRoot          string
	GoPackage          string
	GoModule           string
	ConverterOutputDir string
	HandlersOutputDir  string
	GeneratorFuncs     []GeneratorFunc
}

type ProtoPackage struct {
	name               string
	protoRoot          string
	goPackagePath      string
	goPackageName      string
	goModule           string
	converterOutputDir string
	converterPackage   string
	protoOutputDir     string
	handlersOutputDir  string
	protoPackagePath   string
	generatorFuncs     []GeneratorFunc
	tmpl               *template.Template
	files              []*FileSchema
	converters         convertersData
}

func NewProtoPackage(conf ProtoPackageConfig) *ProtoPackage {
	p := &ProtoPackage{
		name:               conf.Name,
		protoRoot:          conf.ProtoRoot,
		goPackagePath:      conf.GoPackage,
		goModule:           conf.GoModule,
		converterOutputDir: conf.ConverterOutputDir,
		handlersOutputDir:  conf.HandlersOutputDir,
		generatorFuncs:     conf.GeneratorFuncs,
	}

	if conf.Name == "" {
		log.Fatalf("Missing proto package definition.")
	}

	if conf.GoPackage == "" {
		log.Fatalf("Missing go package definition.")
	}

	p.goPackageName = path.Base(conf.GoPackage)

	protoPackagePath := strings.ReplaceAll(conf.Name, ".", "/")
	p.protoPackagePath = protoPackagePath
	p.protoOutputDir = filepath.Join(conf.ProtoRoot, protoPackagePath)

	if p.converterOutputDir == "" {
		p.converterOutputDir = "gen/converter"
	}

	if p.handlersOutputDir == "" {
		p.handlersOutputDir = "gen/handlers"
	}

	tmpl, err := template.New("protoTemplates").Funcs(funcMap).ParseFS(templateFS, "templates/*")
	if err != nil {
		fmt.Print(fmt.Errorf("Failed to initiate tmpl instance for the generator: %w", err))
		os.Exit(1)
	}
	p.tmpl = tmpl

	p.converterPackage = filepath.Base(p.converterOutputDir)
	converters := convertersData{
		Package:   p.converterPackage,
		GoPackage: p.goPackageName,
		Imports:   Set{p.goPackagePath: present}, RepeatedConverters: make(Set),
	}
	p.converters = converters

	return p
}

func (p *ProtoPackage) NewFile(s FileSchema) *FileSchema {
	newFile := &FileSchema{
		Name:       s.Name + ".proto",
		Package:    p,
		Imports:    make(Set),
		Options:    s.Options,
		Extensions: s.Extensions,
		Enums:      s.Enums,
		Messages:   s.Messages,
		Services:   s.Services,
	}
	p.files = append(p.files, newFile)
	return newFile
}

func (p *ProtoPackage) BuildFiles() []FileData {
	out := make([]FileData, len(p.files))
	var fileErrors error

	for _, f := range p.files {
		imports := make(Set)

		file := FileData{
			Package:    f.Package,
			Imports:    imports,
			Extensions: f.Extensions,
			Options:    f.Options,
			Enums:      f.Enums,
		}

		if len(f.Extensions.File)+len(f.Extensions.Service)+len(f.Extensions.Message)+len(f.Extensions.Field)+len(f.Extensions.OneOf) > 0 {
			imports["google/protobuf/descriptor.proto"] = present
		}

		var messageErrors error

		for _, m := range f.Messages {
			var errAgg error

			message, err := m.Build(imports)
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
			fileErrors = errors.Join(fileErrors, indentErrors(fmt.Sprintf("Errors in the file %s", f.Name), messageErrors))
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
