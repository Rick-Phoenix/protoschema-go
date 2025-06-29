// in schemabuilder/protogen/generator.go
package schemabuilder

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"log"
	"maps"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

type protoFileData struct {
	PackageName string
	ServiceData
}

type convertersData struct {
	Package            string
	Imports            Set
	Converters         []messageConverter
	RepeatedConverters Set
}

type ProtoGenerator struct {
	goModule             string
	converterOutputDir   string
	converterPackage     string
	protoPackage         string
	protoOutputDir       string
	handlersOutputDir    string
	protoPackageBasePath string
	protoGenPath         string
	services             []ServiceSchema
	generatorFuncs       []GeneratorFunc
	tmpl                 *template.Template
}

type GeneratorFunc func(s ServiceData) error

type ProtoGeneratorConfig struct {
	GoModule           string
	ConverterOutputDir string
	ProtoPackage       string
	ProtoRoot          string
	ProtoGenPath       string
	HandlersOutputDir  string
	GeneratorFuncs     []GeneratorFunc
}

type ServiceHandler struct {
	ResourceName string
	Handlers     []HandlerData
	Imports      Set
	GenPkg       string
}

//go:embed templates/*
var templateFS embed.FS

func NewProtoGenerator(c ProtoGeneratorConfig) *ProtoGenerator {
	out := &ProtoGenerator{
		protoPackage: c.ProtoPackage,
		goModule:     c.GoModule, converterOutputDir: c.ConverterOutputDir,
		protoGenPath: c.ProtoGenPath, handlersOutputDir: c.HandlersOutputDir,
		generatorFuncs: c.GeneratorFuncs,
	}
	if c.ProtoPackage == "" {
		log.Fatalf("Missing proto package definition.")
	}
	protoPackageBasePath := strings.ReplaceAll(c.ProtoPackage, ".", "/")
	out.protoPackageBasePath = protoPackageBasePath
	out.protoOutputDir = filepath.Join(c.ProtoRoot, protoPackageBasePath)

	if out.converterOutputDir == "" {
		out.converterOutputDir = "gen/converter"
	}

	if out.handlersOutputDir == "" {
		out.handlersOutputDir = "gen/handlers"
	}

	tmpl, err := template.New("protoTemplates").Funcs(funcMap).ParseFS(templateFS, "templates/*")
	if err != nil {
		fmt.Print(fmt.Errorf("Failed to initiate tmpl instance for the generator: %w", err))
		os.Exit(1)
	}
	out.tmpl = tmpl

	out.converterPackage = filepath.Base(out.converterOutputDir)

	return out
}

func (g *ProtoGenerator) Services(services ...ServiceSchema) *ProtoGenerator {
	g.services = services
	return g
}

func (g *ProtoGenerator) BuildServices() []ServiceData {
	out := []ServiceData{}
	var serviceErrors error

	for _, s := range g.services {
		serviceData, err := NewProtoService(s)
		serviceErrors = errors.Join(serviceErrors, indentErrors(fmt.Sprintf("Errors for the service schema %q", s.Resource.Name), err))
		out = append(out, serviceData)
	}

	if serviceErrors != nil {
		fmt.Printf("The following errors occurred:\n\n")
		log.Fatal(serviceErrors)
	}

	return out
}

func (g *ProtoGenerator) getTmpl() *template.Template {
	return g.tmpl
}

func (g *ProtoGenerator) genConnectHandler(s ServiceData) error {
	tmpl := g.tmpl
	var handlerBuffer bytes.Buffer
	handlerData := ServiceHandler{GenPkg: filepath.Base(g.protoGenPath), ResourceName: s.ResourceName, Handlers: s.Handlers, Imports: Set{path.Join(g.goModule, g.protoGenPath): present}}
	if err := tmpl.ExecuteTemplate(&handlerBuffer, "handler.go.tmpl", handlerData); err != nil {
		return fmt.Errorf("Failed to execute template: %w", err)
	}

	handlerOut := filepath.Join(g.handlersOutputDir, strings.ToLower(s.ResourceName)+"_handler.go")

	if err := os.MkdirAll(filepath.Dir(handlerOut), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(handlerOut, handlerBuffer.Bytes(), 0644); err != nil {
		return err
	}

	fmt.Printf("✅ Generated handler in %s\n", handlerOut)

	return nil
}

func (g *ProtoGenerator) Generate() error {
	servicesData := g.BuildServices()
	converters := convertersData{
		Imports: make(Set), RepeatedConverters: make(Set), Package: g.converterPackage,
	}

	tmpl := g.tmpl

	for _, s := range servicesData {
		g.genConnectHandler(s)
		templateData := protoFileData{
			PackageName: g.protoPackage,
			ServiceData: s,
		}

		outputFile := strings.ToLower(s.ResourceName) + ".proto"
		outputPath := filepath.Join(g.protoOutputDir, outputFile)
		delete(s.Imports, filepath.Join(g.protoPackageBasePath, outputFile))

		var outputBuffer bytes.Buffer
		if err := tmpl.ExecuteTemplate(&outputBuffer, "service.proto.tmpl", templateData); err != nil {
			return fmt.Errorf("Failed to execute template: %w", err)
		}

		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(outputPath, outputBuffer.Bytes(), 0644); err != nil {
			return err
		}

		fmt.Printf("✅ Successfully generated proto file at: %s\n", outputPath)

		_, err := exec.LookPath("buf")
		if err != nil {
			fmt.Println("Could not format the generated proto file. Is the buf cli in PATH?")
		} else {
			cmd := exec.Command("buf", "format", "-w", outputPath)

			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				fmt.Printf("Error while attempting to format the file %q: %s\n", outputPath, err.Error())
			}
		}

		if s.Converters != nil {
			converters.Converters = append(converters.Converters, s.Converters.Converters...)
			maps.Copy(converters.Imports, s.Converters.Imports)
			maps.Copy(converters.RepeatedConverters, s.Converters.RepeatedConverters)
		}

		for _, f := range g.generatorFuncs {
			err := f(s)
			if err != nil {
				return err
			}
		}
	}

	if len(converters.Converters) > 0 {
		if g.protoGenPath != "" {
			converters.Imports[path.Join(g.goModule, g.protoGenPath)] = present
		}

		var outputBuffer bytes.Buffer
		if err := tmpl.ExecuteTemplate(&outputBuffer, "converter.go.tmpl", converters); err != nil {
			fmt.Printf("Failed to execute template: %s", err.Error())
		}

		outputPath := filepath.Join(g.converterOutputDir, g.converterPackage+".go")

		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return err
		}

		if err := os.WriteFile(outputPath, outputBuffer.Bytes(), 0644); err != nil {
			fmt.Print(err)
		}

		fmt.Printf("✅ Successfully generated converter at: %s\n", outputPath)
	}

	return nil
}

var funcMap = template.FuncMap{
	"fmtOpt": func(o ProtoOption) string {
		opt, err := getProtoOption(o.Name, o.Value)
		if err != nil {
			fmt.Println(err.Error())
			return "error"
		}

		return "option " + opt
	},
	"join": func(e []string, sep string) string {
		str := ""

		for i, s := range e {
			str += fmt.Sprintf("%q", s)

			if i != len(e)-1 {
				str += ", "
			}
		}

		return str
	},
	"joinInt":   joinIntSlice,
	"joinInt32": joinInt32Slice,
	"joinUint":  joinUintSlice,
	"joinRange": func(r []Range) string {
		str := ""

		for i, v := range r {
			str += fmt.Sprintf("%d to %d", v[0], v[1])
			if i != len(r)-1 {
				str += ", "
			}
		}

		return str
	},
	"dec": func(i int) int {
		return i - 1
	},
	"ternary": func(condition bool, trueVal any, falseVal any) any {
		if condition {
			return trueVal
		} else {
			return falseVal
		}
	},
	"keyword": func(isOptional bool, isRepeated bool) string {
		if isOptional {
			return "optional "
		} else if isRepeated {
			return "repeated "
		}

		return ""
	},
	"serviceSuffix": addServiceSuffix,
	"setContains": func(set Set, key string) bool {
		_, present := set[key]
		return present
	},
}
