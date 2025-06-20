// in schemabuilder/protogen/generator.go
package schemabuilder

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

type Options struct {
	ProtoRoot string
	TmplPath  string
}

type ProtoFileData struct {
	PackageName string
	ProtoService
}

var ProtoTypeMap = map[string]string{
	"string":                    "string",
	"bytes":                     "[]byte",
	"bool":                      "bool",
	"float":                     "float32",
	"double":                    "float64",
	"int32":                     "int32",
	"int64":                     "int64",
	"uint32":                    "uint32",
	"uint64":                    "uint64",
	"sint32":                    "int32",
	"sint64":                    "int64",
	"fixed32":                   "uint32",
	"fixed64":                   "uint64",
	"sfixed32":                  "int32",
	"sfixed64":                  "int64",
	"google.protobuf.Timestamp": "timestamppb.Timestamp",
	"google.protobuf.Duration":  "durationpb.Duration",
	"google.protobuf.FieldMask": "field_mask",
	"google.protobuf.Any":       "any",
}

func Generate(s ProtoService, o Options) error {

	protoPackage := strings.ReplaceAll(path.Dir(s.FileOutput), "/", ".")

	templateData := ProtoFileData{
		PackageName:  protoPackage,
		ProtoService: s,
	}

	funcMap := template.FuncMap{
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

	tmpl, err := template.New(filepath.Base(o.TmplPath)).Funcs(funcMap).ParseFiles(o.TmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var outputBuffer bytes.Buffer
	if err := tmpl.Execute(&outputBuffer, templateData); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
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
