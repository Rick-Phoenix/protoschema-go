// in schemabuilder/protogen/generator.go
package schemabuilder

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

type CompleteServiceData struct {
	Imports     []string
	ServiceName string
	Messages    ProtoMessages
}

type Options struct {
	ProtoRoot   string
	Version     string
	ProjectName string
	TmplPath    string
	ServiceName string
	Imports     map[string]bool
}

type ProtoFileData struct {
	PackageName string
	Imports     []string
	Messages    ProtoMessages
	Resource    string
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

var NullableTypes = map[string]string{
	"double": "google.protobuf.DoubleValue",
	"float":  "google.protobuf.FloatValue",
	"int64":  "google.protobuf.Int64Value",
	"uint64": "google.protobuf.UInt64Value",
	"int32":  "google.protobuf.Int32Value",
	"uint32": "google.protobuf.UInt32Value",
	"bool":   "google.protobuf.BoolValue",
	"string": "google.protobuf.StringValue",
	"byte":   "google.protobuf.BytesValue",
	"byte[]": "google.protobuf.BytesValue",
	"uint8":  "google.protobuf.BytesValue",
}

func Generate(s CompleteServiceData, o Options) error {

	templateData := ProtoFileData{
		PackageName: fmt.Sprintf("%s.%s", o.ProjectName, o.Version),
		Resource:    Capitalize(o.ProjectName),
		Imports:     s.Imports,
		Messages:    s.Messages,
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

	outputPath := filepath.Join(o.ProtoRoot, o.ProjectName, o.Version, s.ServiceName+".proto")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, outputBuffer.Bytes(), 0644); err != nil {
		return err
	}

	fmt.Printf("✅ Successfully generated proto file at: %s\n", outputPath)
	return nil
}
