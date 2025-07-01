package protoschema

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

// A protobuf option.
type ProtoOption struct {
	Name  string
	Value any
}

// A preset of commonly used options.
var Options = struct {
	DisableValidator ProtoOption
	ProtoDeprecated  ProtoOption
	AllowAlias       ProtoOption
}{
	DisableValidator: ProtoOption{Name: "(buf.validate.message).disabled", Value: true},
	ProtoDeprecated:  ProtoOption{Name: "deprecated", Value: true},
	AllowAlias:       ProtoOption{Name: "allow_alias", Value: true},
}

// Rule: uses the protovalidate version of oneof. The differences with the standard oneof implementation are:
// 1. Map and repeated fields are allowed
// 2. If more than one field is populated, it will cause an error
// 3. A field must be explicitely set (its default value is not accepted) to be considered populated
// If required is set to true, at least one field must be set.
// The unpopulated fields will be automatically ignored in terms of validation unless specified otherwise.
func ProtoValidateOneof(required bool, fields ...string) ProtoOption {
	mo := ProtoOption{Name: "(buf.validate.message).oneof"}
	values := make(map[string]any)
	values["fields"] = fields

	if required {
		values["required"] = true
	}

	val, err := formatProtoValue(values)
	if err != nil {
		fmt.Printf("Error while formatting the fields for oneof: %v", err)
	}

	mo.Value = val
	return mo
}

func getOptions(optsMap map[string]any, repeatedOpts []string) ([]string, error) {
	flatOpts := []string{}
	var err error

	optsKeys := slices.Sorted(maps.Keys(optsMap))

	for _, name := range optsKeys {
		value := optsMap[name]

		val, fmtErr := getProtoOption(name, value)

		err = errors.Join(err, fmtErr)

		flatOpts = append(flatOpts, val)
	}

	flatOpts = slices.Concat(flatOpts, repeatedOpts)

	if err != nil {
		return flatOpts, err
	}

	return flatOpts, nil
}

func getProtoOption(name string, value any) (string, error) {
	val, err := formatProtoValue(value)
	if err != nil {
		return "", fmt.Errorf("Error while formatting option %q: %w", name, err)
	}

	return fmt.Sprintf("%s = %s", name, val), nil
}
