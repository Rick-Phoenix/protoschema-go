package schemabuilder

import (
	"fmt"
	"reflect"
)

type TableData map[string]*MethodsOut
type TablesDataType map[string]TableData

var TablesData = TablesDataType{
	"User": TableData{
		"Name": StringCol(&MethodsData{
			Get: &ServiceData{
				Request:  StrValid(),
				Response: StrValid().Extend(StrValid().Required()),
			},
		}),
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

func CreateProto(schemaPtr any) ([]ProtoMessage, error) {
	schemaType := reflect.TypeOf(schemaPtr).Elem()
	schemaName := schemaType.Name()

	var schemaData TableData
	var ok bool

	if schemaData, ok = TablesData[schemaName]; !ok {
		return nil, fmt.Errorf("Could not find the data for the schema %s", schemaName)
	}

	GetRequest := &ProtoMessage{
		Name: "Get" + schemaName + "Request",
	}

	GetResponse := &ProtoMessage{
		Name: "Get" + schemaName + "Response",
	}
	CreateRequest := &ProtoMessage{
		Name: "Create" + schemaName + "Request",
	}
	CreateResponse := &ProtoMessage{
		Name: "Create" + schemaName + "Response",
	}
	UpdateRequest := &ProtoMessage{
		Name: "Update" + schemaName + "Request",
	}
	UpdateResponse := &ProtoMessage{
		Name: "Update" + schemaName + "Response",
	}
	DeleteRequest := &ProtoMessage{
		Name: "Delete" + schemaName + "Request",
	}
	DeleteResponse := &ProtoMessage{
		Name: "Delete" + schemaName + "Response",
	}

	for i := range schemaType.NumField() {
		field := schemaType.Field(i)
		fieldName := field.Name

		if !field.IsExported() {
			fmt.Printf("Ignoring unexported column %s...\n", field.Name)
			continue
		}

		var colInstance *MethodsOut

		if colInstance, ok = schemaData[fieldName]; !ok {
			return nil, fmt.Errorf("Could not find the data for the column %s in the table %s", fieldName, schemaName)
		}

		GetRequest.Fields = append(GetRequest.Fields, ProtoField{
			Name:    fieldName,
			Options: colInstance.Get.Request.Rules})
		GetResponse.Fields = append(GetResponse.Fields, ProtoField{
			Name:    fieldName,
			Options: colInstance.Get.Response.Rules})
		CreateRequest.Fields = append(CreateRequest.Fields, ProtoField{
			Name:    fieldName,
			Options: colInstance.Create.Request.Rules})
		CreateResponse.Fields = append(CreateResponse.Fields, ProtoField{
			Name:    fieldName,
			Options: colInstance.Create.Response.Rules})
		UpdateRequest.Fields = append(UpdateRequest.Fields, ProtoField{
			Name:    fieldName,
			Options: colInstance.Update.Request.Rules})
		UpdateResponse.Fields = append(UpdateResponse.Fields, ProtoField{
			Name:    fieldName,
			Options: colInstance.Update.Response.Rules})
		DeleteRequest.Fields = append(DeleteRequest.Fields, ProtoField{
			Name:    fieldName,
			Options: colInstance.Delete.Request.Rules})
		DeleteResponse.Fields = append(DeleteResponse.Fields, ProtoField{
			Name:    fieldName,
			Options: colInstance.Delete.Response.Rules})

	}

	messages := []ProtoMessage{
		*GetRequest, *GetResponse, *CreateRequest, *CreateResponse, *UpdateRequest, *UpdateResponse, *DeleteRequest, *DeleteResponse,
	}

	return messages, nil
}
