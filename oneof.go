package protoschema

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/labstack/gommon/log"
)

// The oneof's fields. The key is the field number for that field.
type OneofFields map[uint32]FieldBuilder

// A function that gets called with the output of the OneofGroup after it has been processed. Can be defined for the entire package or at the single Oneof level.
type OneofHook func(OneofData) error

// The processed data for the Oneof. Gets passed to the Hook after being generated.
type OneofData struct {
	Name     string
	Fields   []FieldData
	Options  []ProtoOption
	Metadata map[string]any
	Package  *ProtoPackage
	File     *FileSchema
	Message  *MessageSchema
}

// The schema for a protobuf Oneof.This should be created with the constructor from a MessageSchema instance to automatically populate the Package, File and Message fields. It can also be used as a struct to define a Oneof that was not defined by using this library.
type OneofGroup struct {
	Name     string
	Required bool
	Fields   OneofFields
	Options  []ProtoOption
	Package  *ProtoPackage
	File     *FileSchema
	Message  *MessageSchema
	Metadata map[string]any
	Hook     OneofHook
}

// Returns a field with a specific name, causing a fatal error if the field is not found. Modifying this field will modify the original value.
func (of *OneofGroup) GetField(name string) FieldBuilder {
	for _, v := range of.Fields {
		if v.GetName() == name {
			return v
		}
	}
	log.Fatalf("Could not find field %q in oneof %q", name, of.Name)
	return nil
}

func (of *OneofGroup) build(imports Set) (OneofData, error) {
	choicesData := []FieldData{}
	var fieldErr error

	oneofKeys := slices.Sorted(maps.Keys(of.Fields))

	for _, number := range oneofKeys {
		field := of.Fields[number]

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
		Name:     of.Name,
		Options:  options,
		Fields:   choicesData,
		Metadata: of.Metadata,
		Package:  of.Package,
		Message:  of.Message,
		File:     of.File,
	}

	if of.Hook != nil {
		err := of.Hook(out)
		return out, err
	}

	return out, nil
}
