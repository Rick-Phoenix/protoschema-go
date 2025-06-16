// in schemabuilder/protogen/generator.go
package schemabuilder

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"strings"
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
	"string":    "string",
	"int64":     "int64",
	"int32":     "int32",
	"bool":      "bool",
	"byte":      "bytes",
	"byte[]":    "bytes",
	"uint8":     "bytes",
	"timestamp": "google.protobuf.Timestamp",
	"float32":   "float",
	"float64":   "double",
	"uint32":    "uint32",
	"uint64":    "uint64",
}

func Generate(s ProtoService, o Options) error {

	protoPackage := strings.ReplaceAll(path.Dir(s.FileOutput), "/", ".")

	templateData := ProtoFileData{
		PackageName:  protoPackage,
		ProtoService: s,
	}

	funcMap := template.FuncMap{
		"join": strings.Join,
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
