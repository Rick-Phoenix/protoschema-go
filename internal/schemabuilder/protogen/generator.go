// in schemabuilder/protogen/generator.go
package protogen

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Rick-Phoenix/gofirst/internal/schemabuilder"
)

type Options struct {
	ProtoRoot   string
	Version     string
	ProjectName string
	TmplPath    string
}

type ProtoFileData struct {
	PackageName string
	Imports     map[string]bool
	Messages    []MessageData
	Resource    string
}
type MessageData struct {
	Name            string
	Fields          []FieldData
	ReservedNumbers []int
}
type FieldData struct {
	Type    string
	Name    string
	Number  int32
	Options string
}

var ProtoTypeMap = map[string]string{
	"string":    "string",
	"int64":     "int64",
	"int32":     "int32",
	"bool":      "bool",
	"bytes":     "bytes",
	"timestamp": "google.protobuf.Timestamp",
	"updatedAT": "google.protobuf.Timestamp",
}

func Generate(t *schemabuilder.TableBuilder, o Options) error {
	imports := make(map[string]bool)
	serviceNameRaw := t.Name
	serviceName := strings.ToLower(strings.TrimSuffix(serviceNameRaw, "Schema"))
	if serviceName == "" {
		return fmt.Errorf("could not derive service name from type: %s", serviceNameRaw)
	}

	getRequest := &MessageData{Name: fmt.Sprintf("Get%sRequest", schemabuilder.Capitalize(serviceName))}
	getResponse := &MessageData{Name: fmt.Sprintf("Get%sResponse", schemabuilder.Capitalize(serviceName))}

	// 2. GATHER DATA FOR TEMPLATE
	fieldNumber := int32(1) // Start field numbers at 1

	for colName, col := range t.Columns {

		protoData := col.Build()
		rules := protoData.Rules
		if len(rules) > 0 {
			imports["buf/validate/validate.proto"] = true
		}

		if protoData.ColType == "timestamp" {
			imports["google/protobuf/timestamp.proto"] = true
		}

		protoType := ProtoTypeMap[protoData.ColType]

		requests := protoData.Requests

		responses := protoData.Responses

		fieldData := FieldData{
			Type:    protoType,
			Name:    schemabuilder.ToSnakeCase(colName),
			Number:  fieldNumber,
			Options: strings.Join(rules, ", "),
		}

		if slices.Contains(requests, "get") {
			getRequest.Fields = append(getRequest.Fields, fieldData)
		}

		if slices.Contains(responses, "get") {
			getResponse.Fields = append(getResponse.Fields, fieldData)
		}

		fieldNumber++
	}

	templateData := ProtoFileData{
		PackageName: fmt.Sprintf("%s.%s", o.ProjectName, o.Version),
		Resource:    schemabuilder.Capitalize(serviceName),
		Imports:     imports,
		Messages: []MessageData{
			*getRequest, *getResponse,
		},
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

	// 4. WRITE TO FILE
	outputPath := filepath.Join(o.ProtoRoot, o.ProjectName, o.Version, serviceName+".proto")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, outputBuffer.Bytes(), 0644); err != nil {
		return err
	}

	fmt.Printf("✅ Successfully generated proto file at: %s\n", outputPath)
	return nil
}
