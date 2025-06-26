package schemabuilder

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

type ProtoOneOfData struct {
	Name    string
	Choices []ProtoFieldData
	Options []ProtoOption
}

type ProtoOneOfGroup struct {
	name     string
	required bool
	choices  OneofChoicesMap
	options  []ProtoOption
}

type ProtoOneOfBuilder interface {
	Build(imports Set) (ProtoOneOfData, error)
}

type OneofChoicesMap map[uint32]ProtoFieldBuilder

func ProtoOneOf(name string, choices OneofChoicesMap, options ...ProtoOption) *ProtoOneOfGroup {
	return &ProtoOneOfGroup{
		choices: choices, options: options, name: name,
	}
}

func (of *ProtoOneOfGroup) Build(imports Set) (ProtoOneOfData, error) {
	choicesData := []ProtoFieldData{}
	var fieldErr error

	oneofKeys := slices.Sorted(maps.Keys(of.choices))

	for _, number := range oneofKeys {
		field := of.choices[number]

		data, err := field.Build(number, imports)
		fieldErr = errors.Join(fieldErr, err)

		if data.IsMap {
			fieldErr = errors.Join(fieldErr, fmt.Errorf("Cannot use map fields in oneof groups (must be wrapped in a message type first)."))
		}

		if data.Repeated {
			fieldErr = errors.Join(fieldErr, fmt.Errorf("Cannot use repeated fields in oneof groups (must be wrapped in a message type first)."))
		}

		if data.Optional {
			fmt.Printf("Ignoring 'optional' for member %q of oneof group %q...\n", data.Name, of.name)
			data.Optional = false
		}

		choicesData = append(choicesData, data)
	}

	if fieldErr != nil {
		return ProtoOneOfData{}, fieldErr
	}

	return ProtoOneOfData{
		Name: of.name, Options: of.options, Choices: choicesData,
	}, nil
}

func (of *ProtoOneOfGroup) Required() *ProtoOneOfGroup {
	of.options = append(of.options, ProtoOption{
		Name:  "(buf.validate.oneof).required",
		Value: "true",
	})
	return of
}
