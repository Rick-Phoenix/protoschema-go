package schemabuilder

import (
	"errors"
	"fmt"
	"log"
	"maps"
	"reflect"
	"slices"
)

type ProtoFieldsMap map[uint32]ProtoFieldBuilder

type Range [2]int32

type ProtoMessageSchema struct {
	Name            string
	Fields          ProtoFieldsMap
	Oneofs          []ProtoOneOfBuilder
	Enums           []ProtoEnumGroup
	Options         []ProtoOption
	ReservedNumbers []uint
	ReservedRanges  []Range
	ReservedNames   []string
	ReferenceOnly   bool
	ImportPath      string
	Model           any
	ModelIgnore     []string
}

func (s *ProtoMessageSchema) GetFields() map[string]ProtoFieldBuilder {
	out := make(map[string]ProtoFieldBuilder)

	keys := slices.Sorted(maps.Keys(s.Fields))

	for _, k := range keys {
		f := s.Fields[k]

		data := f.GetData()
		out[data.Name] = f
	}

	return out
}

type ProtoMessage struct {
	Name            string
	Fields          []ProtoFieldData
	Oneofs          []ProtoOneOfData
	ReservedNumbers []uint
	ReservedRanges  []Range
	ReservedNames   []string
	Options         []ProtoOption
	Enums           []ProtoEnumGroup
}

func (s *ProtoMessageSchema) GetField(n string) ProtoFieldBuilder {
	for _, f := range s.Fields {
		data := f.GetData()
		if data.Name == n {
			return f
		}
	}

	log.Fatalf("Could not find field %q in schema %q", n, s.Name)
	return nil
}

func (s *ProtoMessageSchema) CheckModel() error {
	model := reflect.TypeOf(s.Model).Elem()
	modelName := model.Name()
	msgFields := s.GetFields()
	var err error

	for i := range model.NumField() {
		field := model.Field(i)
		modelFieldName := field.Tag.Get("json")
		ignore := slices.Contains(s.ModelIgnore, modelFieldName)
		fieldType := field.Type.String()

		if ignore {
			continue
		}

		if pfield, exists := msgFields[modelFieldName]; exists {
			delete(msgFields, modelFieldName)
			goType := pfield.GetGoType()
			fieldName := pfield.GetName()

			if pfield.GetGoType() != fieldType && !slices.Contains(s.ModelIgnore, fieldName) {
				err = errors.Join(err, fmt.Errorf("Expected type %q for field %q, found %q.", fieldType, modelFieldName, goType))
			}
		} else {
			err = errors.Join(err, fmt.Errorf("Column %q not found in the proto schema for %q.", modelFieldName, model))
		}
	}

	if len(msgFields) > 0 {
		for name := range msgFields {
			if !slices.Contains(s.ModelIgnore, name) {
				err = errors.Join(err, fmt.Errorf("Unknown field %q found in the message schema for model %q.", name, modelName))
			}
		}
	}

	if err != nil {
		err = IndentErrors(fmt.Sprintf("Validation errors for model %s", modelName), err)
	}

	return err
}

func NewProtoMessage(s ProtoMessageSchema, imports Set) (ProtoMessage, error) {
	var protoFields []ProtoFieldData
	var fieldsErrors error

	if s.Model != nil {
		err := s.CheckModel()
		if err != nil {
			return ProtoMessage{}, err
		}
	}

	fieldNumbers := slices.Sorted(maps.Keys(s.Fields))

	for _, fieldNr := range fieldNumbers {
		fieldBuilder := s.Fields[fieldNr]
		field, err := fieldBuilder.Build(fieldNr, imports)
		if err != nil {
			fieldsErrors = errors.Join(fieldsErrors, IndentErrors(fmt.Sprintf("Errors for field %s", field.Name), err))
		} else {
			protoFields = append(protoFields, field)
		}
	}

	oneOfs := []ProtoOneOfData{}
	var oneOfErrors error

	for _, oneof := range s.Oneofs {
		data, oneofErr := oneof.Build(imports)

		if oneofErr != nil {
			oneOfErrors = errors.Join(oneOfErrors, IndentErrors(fmt.Sprintf("Errors for oneOf member %s", data.Name), oneofErr))
		}
		oneOfs = append(oneOfs, data)
	}

	if fieldsErrors != nil || oneOfErrors != nil {
		return ProtoMessage{}, errors.Join(fieldsErrors, oneOfErrors)
	}

	return ProtoMessage{Name: s.Name, Fields: protoFields, ReservedNumbers: s.ReservedNumbers, ReservedRanges: s.ReservedRanges, ReservedNames: s.ReservedNames, Options: s.Options, Oneofs: oneOfs, Enums: s.Enums}, nil
}

func MessageRef(name string, importPath string) ProtoMessageSchema {
	return ProtoMessageSchema{Name: name, ReferenceOnly: true, ImportPath: importPath}
}

func ProtoEmpty() ProtoMessageSchema {
	return ProtoMessageSchema{Name: "google.protobuf.Empty", ReferenceOnly: true, ImportPath: "google/protobuf/empty.proto"}
}

var (
	DisableValidator = ProtoOption{Name: "(buf.validate.message).disabled", Value: true}
	ProtoDeprecated  = ProtoOption{Name: "deprecated", Value: true}
)

func ProtoCustomOneOf(required bool, fields ...string) ProtoOption {
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

func MessageCelOption(o CelOption) ProtoOption {
	out := ProtoOption{}

	out.Name = "(buf.validate.message).cel"

	out.Value = GetCelOption(o)

	return out
}

func ReservedNumbers(numbers ...uint) []uint {
	return numbers
}

func ReservedRanges(ranges ...Range) []Range {
	return ranges
}

func ReservedNames(names ...string) []string {
	return names
}
