package schemabuilder

import (
	"fmt"
	"reflect"
	"strings"
)

// The ProcessRules function now needs to be generic to handle the builder interface.
func ProcessRules[T any](builder ColumnBuilder[T]) string {
	if builder == nil {
		return ""
	}

	// First, call Build() to get the final Column object.
	finalColumn := builder.Build()
	rules := finalColumn.Rules

	if len(rules) == 0 {
		return ""
	}
	return fmt.Sprintf("[%s]", strings.Join(rules, ", "))
}

type TableSchema struct {
	NoService   bool
	ServiceName string
}

type UserSchema struct {
	ID          ColumnBuilder[int64]  `bun:"id,pk,autoincrement" json:"id"`
	Name        ColumnBuilder[string] `bun:"name,notnull" json:"user_name"`
	Email       ColumnBuilder[string] `bun:"email,unique"`
	Age         ColumnBuilder[int]
	NoService   bool
	ServiceName string
}

var UserExample = UserSchema{
	// This works because the value returned by StringCol().Required().MinLen(3)
	// is a *StringColumnBuilder, which satisfies the ColumnBuilder[string] interface.
	Name: StringCol().Required().MinLen(3),

	Email: StringCol().Required().Email(),

	Age:         IntCol().GreaterThan(17),
	ServiceName: "User",
}

// The FINAL, CORRECTED version of our unwrapper.
func UnwrapToPlainStruct(richSchemaPtr any) any {
	richValue := reflect.ValueOf(richSchemaPtr)
	richStructValue := richValue.Elem()
	richStructType := richStructValue.Type()

	var plainFields []reflect.StructField

	for i := 0; i < richStructType.NumField(); i++ {
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
