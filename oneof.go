package schemabuilder

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

type ProtoOneofData struct {
	Name    string
	Choices []ProtoFieldData
	Options []ProtoOption
}

type ProtoOneofGroup struct {
	Name       string
	IsRequired bool
	Choices    OneofChoices
	Options    []ProtoOption
}

type OneofBuilder interface {
	Build(imports Set) (ProtoOneofData, error)
}

type OneofChoices map[uint32]ProtoFieldBuilder

func OneOf(name string, choices OneofChoices, options ...ProtoOption) *ProtoOneofGroup {
	return &ProtoOneofGroup{
		Choices: choices, Options: options, Name: name,
	}
}

func (of *ProtoOneofGroup) Build(imports Set) (ProtoOneofData, error) {
	choicesData := []ProtoFieldData{}
	var fieldErr error

	oneofKeys := slices.Sorted(maps.Keys(of.Choices))

	for _, number := range oneofKeys {
		field := of.Choices[number]

		data, err := field.Build(number, imports)
		fieldErr = errors.Join(fieldErr, err)

		if data.IsMap {
			fieldErr = errors.Join(fieldErr, fmt.Errorf("Cannot use map fields in oneof groups (must be wrapped in a message type first)."))
		}

		if data.Repeated {
			fieldErr = errors.Join(fieldErr, fmt.Errorf("Cannot use repeated fields in oneof groups (must be wrapped in a message type first)."))
		}

		if data.Optional {
			fmt.Printf("Ignoring 'optional' for member %q of oneof group %q...\n", data.Name, of.Name)
			data.Optional = false
		}

		choicesData = append(choicesData, data)
	}

	if fieldErr != nil {
		return ProtoOneofData{}, fieldErr
	}

	return ProtoOneofData{
		Name: of.Name, Options: of.Options, Choices: choicesData,
	}, nil
}

func (of *ProtoOneofGroup) Required() *ProtoOneofGroup {
	of.Options = append(of.Options, ProtoOption{
		Name:  "(buf.validate.oneof).required",
		Value: "true",
	})
	return of
}
