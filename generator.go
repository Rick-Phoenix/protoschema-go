// in schemabuilder/protogen/generator.go
package schemabuilder

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type ProtoFileData struct {
	PackageName string
	ProtoService
}

type ProtoGenerator struct {
	packageName string
	outputDir   string
	packageRoot string
}

//go:embed templates/*
var templateFS embed.FS

var funcMap = template.FuncMap{
	"fmtOpt": func(o ProtoOption) string {
		opt, err := GetProtoOption(o.Name, o.Value)
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
}

func NewProtoGenerator(protoRoot, packageName string) *ProtoGenerator {
	packageRoot := strings.ReplaceAll(packageName, ".", "/")
	outputDir := filepath.Join(protoRoot, packageRoot)
	return &ProtoGenerator{
		packageName: packageName, outputDir: outputDir, packageRoot: packageRoot,
	}
}

func (g *ProtoGenerator) Generate(s ProtoService) error {
	templateData := ProtoFileData{
		PackageName:  g.packageName,
		ProtoService: s,
	}

	tmpl, err := template.New("protoTemplates").Funcs(funcMap).ParseFS(templateFS, "templates/*")
	if err != nil {
		return fmt.Errorf("Failed to parse template: %w", err)
	}

	var outputBuffer bytes.Buffer
	if err := tmpl.ExecuteTemplate(&outputBuffer, "service.proto.tmpl", templateData); err != nil {
		return fmt.Errorf("Failed to execute template: %w", err)
	}

	outputFile := strings.ToLower(s.ResourceName) + ".proto"
	outputPath := filepath.Join(g.outputDir, outputFile)
	delete(s.Imports, filepath.Join(g.packageRoot, outputFile))

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, outputBuffer.Bytes(), 0644); err != nil {
		return err
	}

	fmt.Printf("✅ Successfully generated proto file at: %s\n", outputPath)

	cmd := exec.Command("buf", "format", "-w", outputPath)

	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error while attempting to format the file %q: %s\n", outputPath, err.Error())
	}

	return nil
}
