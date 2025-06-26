// in schemabuilder/protogen/generator.go
package schemabuilder

import (
	"bytes"
	"embed"
	"fmt"
	"os"
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

func GenerateProtoFile(s ProtoService, o Options) error {
	protoPackage := strings.ReplaceAll(filepath.Dir(s.FileOutput), "/", ".")

	templateData := ProtoFileData{
		PackageName:  protoPackage,
		ProtoService: s,
	}

	tmpl, err := template.New("protoTemplates").Funcs(funcMap).ParseFS(templateFS, "templates/service.proto.tmpl")
	if err != nil {
		return fmt.Errorf("Failed to parse template: %w", err)
	}

	var outputBuffer bytes.Buffer
	if err := tmpl.ExecuteTemplate(&outputBuffer, "service.proto.tmpl", templateData); err != nil {
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
