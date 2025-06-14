package schemabuilder

import (
	"fmt"
	"log"
	"maps"
	"reflect"
	"slices"
)

type TablesDataType map[string]ProtoServiceOutput

type ServiceData struct {
	Request  ProtoMessage
	Response ProtoMessage
}

type FieldData map[string]*ServiceData

var UserSchema = ProtoMessageSchema{
	Fields: ProtoFields{
		"name": ProtoString(1),
	},
}

type ServicesMap map[string]ProtoServiceSchema

type ServicesData map[string]ProtoServiceOutput

func BuildFinalServicesMap(m ServicesMap) ServicesData {
	out := make(ServicesData)

	for resource, serviceSchema := range m {
		out[resource] = NewProtoService(resource, serviceSchema)
	}
}

// Service must know its name. But it's good to have a map of (db) names to services.
// So I might make a wrapper that takes a map like this and passes the names to the service builders
var TablesData = TablesDataType{
	"User": ProtoServiceSchema{
		Get: &ServiceData{
			Request:  NewProtoMessage(ProtoMessageSchema{}),
			Response: NewProtoMessage(ProtoMessageSchema{}),
		},
	},
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

			appendField := func(fieldInfo *ProtoFieldData, serviceType string) {
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
