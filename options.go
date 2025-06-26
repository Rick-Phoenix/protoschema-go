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
