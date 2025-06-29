package schemabuilder

import (
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
	services           []ServiceSchema
	generatorFuncs     []GeneratorFunc
	tmpl               *template.Template
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

	return p
}

func (p *ProtoPackage) NewMessage(s MessageSchema) MessageSchema {
	s.GoPackageName = p.goPackageName
	s.GoPackagePath = p.goPackagePath
	s.ProtoPackage = p.name
	return s
}
