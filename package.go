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
	FileHook           FileHook
	ServiceHook        ServiceHook
	MessageHook        MessageHook
	OneofHook          OneofHook
	ConverterFunc      ConverterFunc
}

type ProtoPackage struct {
	Name               string
	protoRoot          string
	goPackagePath      string
	goPackageName      string
	goModule           string
	converterOutputDir string
	converterPackage   string
	protoOutputDir     string
	handlersOutputDir  string
	protoPackagePath   string
	fileHook           FileHook
	serviceHook        ServiceHook
	messageHook        MessageHook
	oneofHook          OneofHook
	tmpl               *template.Template
	fileSchemas        []*FileSchema
	Converter          ConverterData
	converterFunc      ConverterFunc
}

func NewProtoPackage(conf ProtoPackageConfig) *ProtoPackage {
	p := &ProtoPackage{
		Name:               conf.Name,
		protoRoot:          conf.ProtoRoot,
		goPackagePath:      conf.GoPackage,
		goModule:           conf.GoModule,
		converterOutputDir: conf.ConverterOutputDir,
		handlersOutputDir:  conf.HandlersOutputDir,
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
	converters := ConverterData{
		Package:   p.converterPackage,
		GoPackage: p.goPackageName,
		Imports:   Set{p.goPackagePath: present}, RepeatedConverters: make(Set),
	}
	p.Converter = converters

	return p
}

func (p *ProtoPackage) NewFile(s FileSchema) *FileSchema {
	newFile := &FileSchema{
		Name:       s.Name + ".proto",
		Package:    p,
		imports:    make(Set),
		Options:    s.Options,
		Extensions: s.Extensions,
		enums:      s.enums,
		messages:   s.messages,
		services:   s.services,
		Hook:       s.Hook,
	}
	if s.Hook == nil {
		s.Hook = p.fileHook
	}
	p.fileSchemas = append(p.fileSchemas, newFile)
	return newFile
}

func (p *ProtoPackage) BuildFiles() []FileData {
	out := []FileData{}
	var fileErrors error

	for _, f := range p.fileSchemas {
		file, err := f.Build()
		if err != nil {
			fileErrors = errors.Join(fileErrors, indentErrors(fmt.Sprintf("Errors in the file %s", f.Name), err))
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
