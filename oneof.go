package schemabuilder

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

type OneofData struct {
	Name    string
	Choices []FieldData
	Options []ProtoOption
}

type OneofGroup struct {
	Name     string
	Required bool
	Choices  OneofChoices
	Options  []ProtoOption
}

type OneofChoices map[uint32]FieldBuilder

func (of OneofGroup) Build(imports Set) (OneofData, error) {
	choicesData := []FieldData{}
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

	options := slices.Clone(of.Options)

	if of.Required {
		options = append(options, ProtoOption{
			Name:  "(buf.validate.oneof).required",
			Value: "true",
		})
	}

	if fieldErr != nil {
		return OneofData{}, fieldErr
	}

	return OneofData{
		Name: of.Name, Options: options, Choices: choicesData,
	}, nil
}
