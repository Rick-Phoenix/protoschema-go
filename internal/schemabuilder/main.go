package schemabuilder

import (
	"fmt"
	"reflect"
)

type TableData map[string]ColumnBuilder
type TablesDataType map[string]TableData

type ServiceData struct {
	Request  ColumnBuilder
	Response ColumnBuilder
}

var MethodsData struct {
	Create, Get, Update, Delete *ServiceData
}

var TablesData = TablesDataType{
	"User": TableData{
		"Name":      StringCol().Required().MinLen(3).Requests("create").Responses("get", "create").Extend(StringCol().Required()),
		"Age":       Int64Col().Responses("get").Nullable(),
		"Blob":      BytesCol().Requests("get"),
		"CreatedAt": TimestampCol().Responses("get"),
	},
}

func CreateProto(schemaPtr any) ([]Column, error) {
	schemaType := reflect.TypeOf(schemaPtr).Elem()
	schemaName := schemaType.Name()

	var columns []Column
	var schemaData TableData
	var ok bool

	if schemaData, ok = TablesData[schemaName]; !ok {
		return nil, fmt.Errorf("Could not find the data for the schema %s", schemaName)
	}

	for i := range schemaType.NumField() {
		field := schemaType.Field(i)
		fieldName := field.Name

		if !field.IsExported() {
			fmt.Printf("Ignoring unexported column %s...\n", field.Name)
			continue
		}

		var colInstance ColumnBuilder

		if colInstance, ok = schemaData[fieldName]; !ok {
			return nil, fmt.Errorf("Could not find the data for the column %s in the table %s", fieldName, schemaName)
		}

		colData := colInstance.Build()

		columns = append(columns, colData)

	}

	return columns, nil
}
