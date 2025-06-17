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
	// Scalar Protobuf Types
	"string":   "string",
	"bytes":    "bytes",
	"bool":     "bool",
	"float":    "float",
	"double":   "double",
	"int32":    "int32",
	"int64":    "int64",
	"uint32":   "uint32",
	"uint64":   "uint64",
	"sint32":   "sint32",
	"sint64":   "sint64",
	"fixed32":  "fixed32",
	"fixed64":  "fixed64",
	"sfixed32": "sfixed32",
	"sfixed64": "sfixed64",
	// Well-Known Types
	"google.protobuf.Timestamp": "timestamp",
	"google.protobuf.Duration":  "duration",
	"google.protobuf.FieldMask": "field_mask",
	"google.protobuf.Any":       "any",
	// Note: For other well-known types (like Struct, Value, Wrappers),
	// or custom message/enum types, `protovalidate` might use different
	// rule categories (e.g., `message`, `map`, `repeated`, `wrapper`, `enum`).
	// If you plan to add `const` validation for enums, you would map
	// your enum's proto type name (e.g., `com.example.MyEnum`) to "enum".
}

func Generate(s ProtoService, o Options) error {

	protoPackage := strings.ReplaceAll(path.Dir(s.FileOutput), "/", ".")

	templateData := ProtoFileData{
		PackageName:  protoPackage,
		ProtoService: s,
	}

	funcMap := template.FuncMap{
		"joinInt": JoinIntSlice,
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
