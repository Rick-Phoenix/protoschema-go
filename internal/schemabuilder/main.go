package schemabuilder

import (
	"fmt"
	"reflect"
)

type TableData map[string]ColumnBuilder
type TablesDataType map[string]TableData

var TablesData = TablesDataType{
	"User": TableData{
		"Name":      StringCol().Required().MinLen(3).Requests("create").Responses("get", "create").Nullable(),
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

// The FINAL, CORRECTED version of our unwrapper.
func UnwrapToPlainStruct(richSchemaPtr any) any {
	richValue := reflect.ValueOf(richSchemaPtr)
	richStructValue := richValue.Elem()
	richStructType := richStructValue.Type()

	var plainFields []reflect.StructField

	for i := range richStructType.NumField() {
		richField := richStructType.Field(i) // The field definition from UserSchema

		// richField.Type is now our ColumnBuilder[T] interface.
		// Let's inspect it to find T.
		builderInterfaceType := richField.Type

		// 1. Get the 'Build' method from the interface definition.
		buildMethod, ok := builderInterfaceType.MethodByName("Build")
		if !ok {
			// This field isn't a valid ColumnBuilder, skip it.
			continue
		}

		// 2. The 'Build' method returns one value: a Column[T].
		// We get the type of that return value.
		columnStructType := buildMethod.Type.Out(0) // Out(0) is the first return type

		// 3. Now we have the type for Column[T]. We can inspect its 'Value' field
		// to finally get the type for T.
		valueField, ok := columnStructType.FieldByName("Value")
		if !ok {
			panic("The final Column[T] struct must have a 'Value' field of type T.")
		}
		unwrappedType := valueField.Type // This is T!

		plainFields = append(plainFields, reflect.StructField{
			Name: richField.Name,
			Type: unwrappedType,
			Tag:  richField.Tag, // Read the tag from the UserSchema field
		})
	}

	plainStructType := reflect.StructOf(plainFields)
	return reflect.New(plainStructType).Interface()
}
