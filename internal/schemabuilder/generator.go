// in schemabuilder/protogen/generator.go
package schemabuilder

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

type Options struct {
	ProtoRoot string
}

type ProtoFileData struct {
	PackageName string
	ProtoService
}

//go:embed templates/*
var templateFS embed.FS

func GenerateProtoFile(s ProtoService, o Options) error {
	protoPackage := strings.ReplaceAll(path.Dir(s.FileOutput), "/", ".")
	templatePath := "templates/service.proto.tmpl"

	templateData := ProtoFileData{
		PackageName:  protoPackage,
		ProtoService: s,
	}

	tmpl, err := template.ParseFS(templateFS, "templates/service.proto.tmpl").Funcs(funcMap).ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("Failed to parse template: %w", err)
	}

	var outputBuffer bytes.Buffer
	if err := tmpl.Execute(&outputBuffer, templateData); err != nil {
		return fmt.Errorf("Failed to execute template: %w", err)
	}

	outputPath := filepath.Join(o.ProtoRoot, s.FileOutput)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, outputBuffer.Bytes(), 0644); err != nil {
		return err
	}

	fmt.Printf("✅ Successfully generated proto file at: %s\n", outputPath)
	return nil
}

var funcMap = template.FuncMap{
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
	"joinInt":   JoinIntSlice,
	"joinInt32": JoinInt32Slice,
	"joinUint":  JoinUintSlice,
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
}
