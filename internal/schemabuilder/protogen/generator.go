// in schemabuilder/protogen/generator.go
package protogen

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"

	"github.com/Rick-Phoenix/gofirst/internal/schemabuilder"
)

// Options struct remains the same
type Options struct {
	OutputRoot string
	Version    string
}

// --- Template Data Structs ---
type ProtoFileData struct {
	PackageName string
	Imports     []string
	Messages    []MessageData
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

func extractRules(builder any) []string {
	if builder == nil {
		return nil
	}

	// The type switch is the correct and safe way to handle this.
	switch b := builder.(type) {
	case schemabuilder.ColumnBuilder[string]:
		return b.Build().Rules
	case schemabuilder.ColumnBuilder[int]:
		return b.Build().Rules
	case schemabuilder.ColumnBuilder[int64]:
		return b.Build().Rules
	// As you add new builders (e.g., for bool), you add one line here.
	default:
		// This should ideally not happen if your schema is well-formed.
		fmt.Printf("Warning: unknown builder type encountered: %T\n", b)
		return nil
	}
}

func Generate(schemaPtr any, tmplPath, outputRoot, version string) error {
	schemaType := reflect.TypeOf(schemaPtr).Elem()

	// 1. DERIVE SERVICE NAME
	serviceNameRaw := schemaType.Name()
	serviceName := strings.ToLower(strings.TrimSuffix(serviceNameRaw, "Schema"))
	if serviceName == "" {
		return fmt.Errorf("could not derive service name from type: %s", serviceNameRaw)
	}

	// 2. GATHER DATA FOR TEMPLATE
	fieldNumber := int32(1) // Start field numbers at 1
	var fields []FieldData

	for i := range schemaType.NumField() {
		fieldDef := schemaType.Field(i)
		if !unicode.IsUpper(rune(fieldDef.Name[0])) {
			continue // Skip unexported fields
		}

		builderInstance := reflect.ValueOf(schemaPtr).Elem().Field(i).Interface()
		rules := extractRules(builderInstance)
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

		protoType := "string" // Default
		if goType.Kind() == reflect.Int64 {
			protoType = "int64"
		}

		fields = append(fields, FieldData{
			Type:    protoType,
			Name:    strings.ToLower(fieldDef.Name), // Convert to snake_case
			Number:  fieldNumber,
			Options: strings.Join(rules, ", "),
		})
		fieldNumber++
	}

	templateData := ProtoFileData{
		PackageName: fmt.Sprintf("myapp.%s.%s", serviceName, version),
		Imports:     []string{"buf/validate/validate.proto"},
		Messages: []MessageData{
			{Name: strings.ToTitle(serviceName), Fields: fields}, // e.g., message User
		},
	}

	funcMap := template.FuncMap{
		"join": strings.Join,
	}

	tmpl, err := template.New(filepath.Base(tmplPath)).Funcs(funcMap).ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var outputBuffer bytes.Buffer
	if err := tmpl.Execute(&outputBuffer, templateData); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// 4. WRITE TO FILE
	outputPath := filepath.Join(outputRoot, version, serviceName+".proto")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, outputBuffer.Bytes(), 0644); err != nil {
		return err
	}

	fmt.Printf("✅ Successfully generated proto file at: %s\n", outputPath)
	return nil
}
