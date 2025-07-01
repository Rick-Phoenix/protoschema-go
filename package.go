package schemabuilder

import (
	"errors"
	"fmt"
	"log"
	"maps"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

type ProtoPackageConfig struct {
	Name               string
	BasePath           string
	ProtoRoot          string
	GoPackage          string
	GoModule           string
	ConverterOutputDir string
	FileHook           FileHook
	ServiceHook        ServiceHook
	MessageHook        MessageHook
	OneofHook          OneofHook
	ConverterFunc      ConverterFunc
}

type ProtoPackage struct {
	Name               string
	GoPackagePath      string
	GoPackageName      string
	protoRoot          string
	goModule           string
	converterOutputDir string
	converterPackage   string
	protoOutputDir     string
	protoPackagePath   string
	fileHook           FileHook
	serviceHook        ServiceHook
	messageHook        MessageHook
	oneofHook          OneofHook
	tmpl               *template.Template
	fileSchemas        []*FileSchema
	converter          converterData
	converterFunc      ConverterFunc
}

func (p *ProtoPackage) GetName() string {
	if p == nil {
		return ""
	}

	return p.Name
}

func (p *ProtoPackage) GetGoPackageName() string {
	if p == nil {
		return ""
	}

	if p.GoPackageName == "" {
		if goPkgBase := path.Base(p.GoPackagePath); goPkgBase != "." {
			return goPkgBase
		}
	}

	return p.GoPackageName
}

func (p *ProtoPackage) GetBasePath() string {
	if p == nil {
		return ""
	}

	if p.protoPackagePath == "" {
		pkgPath := strings.ReplaceAll(p.Name, ".", "/")
		p.protoPackagePath = pkgPath
		return pkgPath
	}

	return p.protoPackagePath
}

func (p *ProtoPackage) GetGoPackagePath() string {
	if p == nil {
		return ""
	}

	return p.GoPackagePath
}

func NewProtoPackage(conf ProtoPackageConfig) *ProtoPackage {
	p := &ProtoPackage{
		Name:               conf.Name,
		protoRoot:          conf.ProtoRoot,
		GoPackagePath:      conf.GoPackage,
		goModule:           conf.GoModule,
		converterOutputDir: conf.ConverterOutputDir,
		fileHook:           conf.FileHook,
		serviceHook:        conf.ServiceHook,
		messageHook:        conf.MessageHook,
		oneofHook:          conf.OneofHook,
		converterFunc:      conf.ConverterFunc,
		protoPackagePath:   conf.BasePath,
	}

	if conf.Name == "" {
		log.Fatalf("Missing proto package definition.")
	}

	if conf.GoPackage == "" {
		log.Fatalf("Missing go package definition.")
	}

	p.GoPackageName = path.Base(conf.GoPackage)

	if p.protoPackagePath == "" {
		p.protoPackagePath = strings.ReplaceAll(conf.Name, ".", "/")
	}
	p.protoOutputDir = filepath.Join(conf.ProtoRoot, p.protoPackagePath)

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

func (p *ProtoPackage) NewFile(s FileSchema) *FileSchema {
	newFile := &FileSchema{
		Name:       s.Name + ".proto",
		Package:    p,
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
