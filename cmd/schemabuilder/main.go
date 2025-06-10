package main

import (
	"fmt"
	"reflect"

	sb "github.com/Rick-Phoenix/gofirst/internal/schemabuilder"
)

func main() {
	fmt.Println("--- Step 1: Defining our rich schema instance ---")

	fmt.Printf("Original rich schema type: %T\n\n", sb.UserExample)

	fmt.Println("--- Step 2: Calling UnwrapToPlainStruct ---")
	plainUser := sb.UnwrapToPlainStruct(sb.UserExample)
	fmt.Printf("The returned plain struct is of type: %T\n", plainUser)
	fmt.Printf("Value of the plain struct instance: %#v\n\n", plainUser)

	fmt.Println("--- Step 3: Verifying the new struct's fields and tags ---")
	// To inspect the result, we use reflection again!
	// We use .Elem() because plainUser is a pointer to a struct.
	plainType := reflect.TypeOf(plainUser).Elem()

	for i := range plainType.NumField() {
		field := plainType.Field(i)
		fmt.Printf(
			"Field %d: Name=%s, Type=%s, Tag='%s'\n",
			i,
			field.Name,
			field.Type,
			field.Tag,
		)
	}

	rules := sb.ProcessRules(sb.UserExample.Name)

	fmt.Print(rules)
}
