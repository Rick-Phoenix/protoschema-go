package schemabuilder

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

type OneofChoices map[uint32]FieldBuilder

type OneofHook func(OneofData) error

type OneofData struct {
	Name     string
	Choices  []FieldData
	Options  []ProtoOption
	Metadata map[string]any
	Hook     OneofHook
}

type OneofGroup struct {
	Name     string
	Required bool
	Choices  OneofChoices
	Options  []ProtoOption
	Package  *ProtoPackage
	File     *FileSchema
	Message  *MessageSchema
	Metadata map[string]any
	Hook     OneofHook
}

func (of OneofGroup) GetField(name string) FieldBuilder {
	for _, v := range of.Choices {
		if v.GetName() == name {
			return v
		}
	}
	fmt.Printf("Could not find field %q in oneof %q", name, of.Name)
	return nil
}

func (of OneofGroup) GetFields() []FieldData {
	data := make([]FieldData, len(of.Choices))

	i := 0
	for _, field := range of.Choices {
		data[i] = field.GetData()
		i++
	}

	return data
}

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

	out := OneofData{
		Name: of.Name, Options: options, Choices: choicesData,
	}

	if of.Hook != nil {
		err := of.Hook(out)
		return out, err
	}

	return out, nil
}
