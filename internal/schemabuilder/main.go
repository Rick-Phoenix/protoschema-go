package schemabuilder

import (
	"fmt"
	"log"
	"maps"
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

type CompleteServiceData struct {
	Imports     []string
	ServiceName string
	Messages    ProtoMessages
}

type FieldData map[string]*ServiceData

var TablesData = TablesDataType{
	"User": TableData{
		"Name": FieldData{
			"Get": &ServiceData{
				Request:  ProtoString(1),
				Response: ProtoString(2).Extend(ProtoString(0).Required()),
			},
		},
	},
}

type ProtoMessage struct {
	Name     string
	Fields   []ProtoField
	Reserved []int
}

type ProtoField struct {
	Name    string
	Type    string
	Options map[string]string
}

var ValidMethods = []string{"Get", "Create", "Update", "Delete"}

type ProtoMessages map[string]*ProtoMessage

func CreateProto(schemaPtr any) (*CompleteServiceData, error) {
	schemaType := reflect.TypeOf(schemaPtr).Elem()
	schemaName := schemaType.Name()

	var schemaData TableData
	var ok bool

	if schemaData, ok = TablesData[schemaName]; !ok {
		return nil, fmt.Errorf("Could not find the data for the schema %s", schemaName)
	}

	imports := make(map[string]bool)

	imports["buf/validate/validate.proto"] = true

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

			appendField := func(fieldInfo *Column, serviceType string) {
				messageName := methodName + schemaName + serviceType
				coltype := fieldInfo.ColType
				if coltype != fieldType {
					log.Fatalf("Mismatch in the type")
				}

				switch coltype {
				case "timestamp":
					imports["google/protobuf/timestamp.proto"] = true
				case "fieldMask":
					imports["google/protobuf/field_mask.proto"] = true
				}

				var protoType string

				if fieldType == "sql.NullString" || fieldInfo.Nullable == true {
					imports["google/protobuf/wrappers.proto"] = true
					protoType = NullableTypes[coltype]
				} else {
					protoType = ProtoTypeMap[coltype]
				}

				if _, ok := messages[messageName]; !ok {
					messages[messageName] = &ProtoMessage{Name: messageName}
				}

				messages[messageName].Fields = append(messages[messageName].Fields, ProtoField{Name: fieldName, Options: fieldInfo.Rules, Type: protoType})
			}

			if methodInstructions.Request != nil {
				fieldInfo := methodInstructions.Request.Build()
				appendField(&fieldInfo, "Request")
			}

			if methodInstructions.Response != nil {
				fieldInfo := methodInstructions.Response.Build()
				appendField(&fieldInfo, "Response")
			}
		}

	}

	completeServiceData := &CompleteServiceData{
		ServiceName: schemaName, Imports: slices.Collect(maps.Keys(imports)), Messages: messages,
	}
	return completeServiceData, nil
}
