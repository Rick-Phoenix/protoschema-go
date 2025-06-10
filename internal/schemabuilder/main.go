package main

import (
	"fmt"
	"reflect"
	"strings"
)

// Column is the final data structure. It's simple and holds the results.
type Column[T any] struct {
	Value T // For our UnwrapToPlainStruct helper
	Rules []string
}

// ColumnBuilder is an interface for any type that can produce a Column[T].
type ColumnBuilder[T any] interface {
	// Any type that has this method signature automatically implements the interface.
	Build() Column[T]
}

// StringColumnBuilder is a temporary object used to build a Column[string].
type StringColumnBuilder struct {
	rules []string
}

// StringCol is our "constructor" function. It's the entry point.
// It returns a pointer to the builder so we can chain methods.
func StringCol() *StringColumnBuilder {
	return &StringColumnBuilder{}
}

// --- Validation Methods ---

// Required adds the 'required' rule and returns the builder for chaining.
func (b *StringColumnBuilder) Required() *StringColumnBuilder {
	b.rules = append(b.rules, "(buf.validate.field).required = true")
	return b // Return self
}

// MinLen adds a minimum length rule.
func (b *StringColumnBuilder) MinLen(len uint64) *StringColumnBuilder {
	rule := fmt.Sprintf("(buf.validate.field).string.min_len = %d", len)
	b.rules = append(b.rules, rule)
	return b
}

// Email adds the 'email' format rule.
func (b *StringColumnBuilder) Email() *StringColumnBuilder {
	rule := fmt.Sprintf("(buf.validate.field).string.email = true")
	b.rules = append(b.rules, rule)
	return b
}

func (b *StringColumnBuilder) Build() Column[string] {
	return Column[string]{Rules: b.rules}
}

type IntColumnBuilder struct {
	rules []string
}

func IntCol() *IntColumnBuilder { return &IntColumnBuilder{} }
func (b *IntColumnBuilder) GreaterThan(val int64) *IntColumnBuilder {
	b.rules = append(b.rules, fmt.Sprintf("(buf.validate.field).int64.gt = %d", val))
	return b
}
func (b *IntColumnBuilder) Build() Column[int] {
	return Column[int]{Rules: b.rules}
}

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

type UserSchema struct {
	ID    ColumnBuilder[int64]  `bun:"id,pk,autoincrement" json:"id"`
	Name  ColumnBuilder[string] `bun:"name,notnull" json:"user_name"`
	Email ColumnBuilder[string] `bun:"email,unique"`
	Age   ColumnBuilder[int]
}

var UserExample = UserSchema{
	// This works because the value returned by StringCol().Required().MinLen(3)
	// is a *StringColumnBuilder, which satisfies the ColumnBuilder[string] interface.
	Name: StringCol().Required().MinLen(3),

	Email: StringCol().Required().Email(),

	Age: IntCol().GreaterThan(17),
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

func main() {
	fmt.Println("--- Step 1: Defining our rich schema instance ---")

	fmt.Printf("Original rich schema type: %T\n\n", &UserExample)

	fmt.Println("--- Step 2: Calling UnwrapToPlainStruct ---")
	plainUser := UnwrapToPlainStruct(&UserExample)
	fmt.Printf("The returned plain struct is of type: %T\n", plainUser)
	fmt.Printf("Value of the plain struct instance: %#v\n\n", plainUser)

	fmt.Println("--- Step 3: Verifying the new struct's fields and tags ---")
	// To inspect the result, we use reflection again!
	// We use .Elem() because plainUser is a pointer to a struct.
	plainType := reflect.TypeOf(plainUser).Elem()

	for i := 0; i < plainType.NumField(); i++ {
		field := plainType.Field(i)
		fmt.Printf(
			"Field %d: Name=%s, Type=%s, Tag='%s'\n",
			i,
			field.Name, // The field's name
			field.Type, // The field's Go type
			field.Tag,  // The field's struct tag
		)
	}

	rules := ProcessRules(UserExample.Name)

	fmt.Print(rules)
}
