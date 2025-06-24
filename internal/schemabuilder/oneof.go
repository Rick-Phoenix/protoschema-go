package schemabuilder

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
)

type ProtoOneOfData struct {
	Name    string
	Choices []ProtoFieldData
	Options []ProtoOption
	Imports []string
}

type ProtoOneOfGroup struct {
	required bool
	choices  OneofChoicesMap
	options  []ProtoOption
}

type ProtoOneOfBuilder interface {
	Build(name string, imports Set) (ProtoOneOfData, error)
}

type ProtoOneofsMap map[string]ProtoOneOfBuilder

type OneofChoicesMap map[string]ProtoFieldBuilder

func ProtoOneOf(choices OneofChoicesMap, options ...ProtoOption) *ProtoOneOfGroup {
	return &ProtoOneOfGroup{
		choices: choices, options: options,
	}
}

func (of *ProtoOneOfGroup) Build(name string, imports Set) (ProtoOneOfData, error) {
	choicesData := []ProtoFieldData{}
	var fieldErr error

	oneofKeys := slices.Sorted(maps.Keys(of.choices))

	for _, name := range oneofKeys {
		field := of.choices[name]

		data, err := field.Build(0, imports)

		fieldErr = errors.Join(fieldErr, err)

		if data.Optional {
			fmt.Printf("Ignoring 'optional' for member '%s' of a oneof group...\n", name)
		}

		if data.Repeated {
			fieldErr = errors.Join(fieldErr, fmt.Errorf("Cannot use a repeated field inside a oneof group."))
		}

		if strings.HasPrefix(data.ProtoType, "map<") {
			fieldErr = errors.Join(fieldErr, fmt.Errorf("Cannot use a map field inside a oneof group."))
		}

		choicesData = append(choicesData, data)

	}

	if fieldErr != nil {
		return ProtoOneOfData{}, fieldErr
	}

	return ProtoOneOfData{
		Name: name, Options: of.options, Choices: choicesData,
	}, nil
}

func (of *ProtoOneOfGroup) Required() *ProtoOneOfGroup {
	of.options = append(of.options, ProtoOption{
		Name:  "(buf.validate.oneof).required",
		Value: "true",
	})
	return of
}
