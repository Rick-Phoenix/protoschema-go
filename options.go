package schemabuilder

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

type ProtoOption struct {
	Name  string
	Value any
}

var Options = struct {
	DisableValidator ProtoOption
	ProtoDeprecated  ProtoOption
	AllowAlias       ProtoOption
}{
	DisableValidator: ProtoOption{Name: "(buf.validate.message).disabled", Value: true},
	ProtoDeprecated:  ProtoOption{Name: "deprecated", Value: true},
	AllowAlias:       ProtoOption{Name: "allow_alias", Value: true},
}

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
