// in schemabuilder/protogen/generator.go
package protogen

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/Rick-Phoenix/gofirst/internal/schemabuilder"
)

// Options struct remains the same
type Options struct {
	ProtoRoot   string
	Version     string
	ProjectName string
	TmplPath    string
}

// --- Template Data Structs ---
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

type ProtoData struct {
	Rules []string

	Requests []string

	Responses []string
}

func extractProtoData(builder any) *ProtoData {

	// The type switch is the correct and safe way to handle this.
	switch b := builder.(type) {
	case schemabuilder.ColumnBuilder[string]:
		data := b.Build()
		return &ProtoData{Rules: data.Rules, Responses: data.Responses, Requests: data.Requests}
	case schemabuilder.ColumnBuilder[int]:
		data := b.Build()
		return &ProtoData{Rules: data.Rules, Responses: data.Responses, Requests: data.Requests}

	case schemabuilder.ColumnBuilder[int64]:
		data := b.Build()
		return &ProtoData{Rules: data.Rules, Responses: data.Responses, Requests: data.Requests}

	case schemabuilder.ColumnBuilder[time.Time]:
		data := b.Build()
		return &ProtoData{Rules: data.Rules, Responses: data.Responses, Requests: data.Requests}

	default:
		// This should ideally not happen if your schema is well-formed.
		fmt.Printf("Warning: unknown builder type encountered: %T\n", b)
		return nil
	}

}

func getProtoType(t reflect.Type, i *map[string]bool) string {
	var protoType string

	typeKind := t.Kind()
	switch typeKind {
	case reflect.String:
		protoType = "string"
	case reflect.Bool:
		protoType = "bool"
	case reflect.Int32: // Assuming you will add Int32Col builder later
		protoType = "int32"
	case reflect.Int64:
		protoType = "int64"
	case reflect.Float32: // Assuming Float32Col builder
		protoType = "float"
	case reflect.Float64: // Assuming Float64Col builder
		protoType = "double"
	case reflect.Slice:
		// Check for []byte specifically
		if t.Elem().Kind() == reflect.Uint8 && t.Elem().Name() == "byte" {
			protoType = "bytes"
		}
	}

	if protoType == "" {
		typePkgPath := t.PkgPath()
		typePkgName := t.Name()

		if typePkgPath == "time" && typePkgName == "Time" {
			protoType = "google.protobuf.Timestamp"
			(*i)["google/protobuf/timestamp.proto"] = true
		}
	}

	if protoType == "" {
		log.Fatalf("Could not determine a proto type for %s", t.Name())
	}

	return protoType
}

func Generate(schemaPtr any, o Options) error {
	imports := make(map[string]bool)
	schemaType := reflect.TypeOf(schemaPtr).Elem()
	// 1. DERIVE SERVICE NAME
	serviceNameRaw := schemaType.Name()
	serviceName := strings.ToLower(strings.TrimSuffix(serviceNameRaw, "Schema"))
	if serviceName == "" {
		return fmt.Errorf("could not derive service name from type: %s", serviceNameRaw)
	}

	getRequest := &MessageData{Name: fmt.Sprintf("Get%sRequest", schemabuilder.Capitalize(serviceName))}
	getResponse := &MessageData{Name: fmt.Sprintf("Get%sResponse", schemabuilder.Capitalize(serviceName))}

	// 2. GATHER DATA FOR TEMPLATE
	fieldNumber := int32(1) // Start field numbers at 1

	for i := range schemaType.NumField() {
		fieldDef := schemaType.Field(i)
		fmt.Println(fieldDef)

		builderInstance := reflect.ValueOf(schemaPtr).Elem().Field(i).Interface()
		protoData := extractProtoData(builderInstance)
		rules := protoData.Rules
		if len(rules) > 0 {
			imports["buf/validate/validate.proto"] = true
		}
		builderValue := reflect.ValueOf(builderInstance)

		// 2. Find and Call the "Build" method using reflection.
		//    It returns a slice of reflect.Value, one for each return value.
		buildResults := builderValue.MethodByName("Build").Call(nil)

		// 3. Get the first return value, which is our Column[T] struct.
		finalColumnValue := buildResults[0]

		// 4. Get the 'Value' field from that struct.
		valueField := finalColumnValue.FieldByName("Value")

		// 5. Get the Go type of that field. This is T!
		goType := valueField.Type()

		protoType := getProtoType(goType, &imports)

		requests := protoData.Requests

		responses := protoData.Responses

		fieldData := FieldData{
			Type:    protoType,
			Name:    schemabuilder.ToSnakeCase(fieldDef.Name),
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
