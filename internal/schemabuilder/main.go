package schemabuilder

import (
	"fmt"
	"log"
	"reflect"
	"slices"
)

type TableData map[string]FieldData
type TablesDataType map[string]TableData

type ServiceData struct {
	Request  ColumnBuilder
	Response ColumnBuilder
}
type MethodsData struct {
	Create, Get, Update, Delete *ServiceData
}

type FieldData map[string]*ServiceData

var TablesData = TablesDataType{
	"User": TableData{
		"Name": FieldData{
			"Get": &ServiceData{
				Request:  StrValid(),
				Response: StrValid().Extend(StrValid().Required()),
			},
		},
	},
}

type ProtoMessage struct {
	Name   string
	Fields []ProtoField
}

type ProtoField struct {
	Name    string
	Type    string
	Options map[string]string
}

var ValidMethods = []string{"Get", "Create", "Update", "Delete"}

func CreateProto(schemaPtr any) (map[string]*ProtoMessage, error) {
	schemaType := reflect.TypeOf(schemaPtr).Elem()
	schemaName := schemaType.Name()

	var schemaData TableData
	var ok bool

	if schemaData, ok = TablesData[schemaName]; !ok {
		return nil, fmt.Errorf("Could not find the data for the schema %s", schemaName)
	}

	var messages = make(map[string]*ProtoMessage)

	for i := range schemaType.NumField() {
		dbColumn := schemaType.Field(i)
		fieldName := dbColumn.Name
		fieldType := dbColumn.Type.String()

		if !dbColumn.IsExported() {
			fmt.Printf("Ignoring unexported column %s...\n", dbColumn.Name)
			continue
		}

		var fieldDefinitions FieldData

		if fieldDefinitions, ok = schemaData[fieldName]; !ok {
			return nil, fmt.Errorf("Could not find the data for the column %s in the table %s", fieldName, schemaName)
		}

		for methodName, methodInstructions := range fieldDefinitions {
			if !slices.Contains(ValidMethods, methodName) {
				log.Fatalf("Invalid method name, %s", methodName)
			}

			if methodInstructions.Request != nil {
				fieldInfo := methodInstructions.Request.Build()
				if fieldInfo.ColType != fieldType {
					log.Fatalf("Mismatch in the type")
				}
				messageName := methodName + schemaName + "Request"
				if _, ok := messages[messageName]; !ok {
					messages[messageName] = &ProtoMessage{Name: messageName}
				}

				messages[messageName].Fields = append(messages[messageName].Fields, ProtoField{Name: fieldName, Options: fieldInfo.Rules})
			}

			if methodInstructions.Response != nil {
				fieldInfo := methodInstructions.Response.Build()
				if fieldInfo.ColType != fieldType {
					log.Fatalf("Mismatch in the type")
				}
				messageName := methodName + schemaName + "Response"
				if _, ok := messages[messageName]; !ok {
					messages[messageName] = &ProtoMessage{Name: messageName}
				}

				messages[messageName].Fields = append(messages[messageName].Fields, ProtoField{Name: fieldName, Options: fieldInfo.Rules})
			}
		}

	}

	return messages, nil
}
