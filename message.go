package schemabuilder

import (
	"errors"
	"fmt"
	"log"
	"maps"
	"reflect"
	"slices"
)

type FieldsMap map[uint32]FieldBuilder

type Range [2]int32

type MessageSchema struct {
	Name            string
	Fields          FieldsMap
	Oneofs          []OneofGroup
	Enums           []EnumGroup
	Options         []ProtoOption
	Messages        []MessageSchema
	ReservedNumbers []uint
	ReservedRanges  []Range
	ReservedNames   []string
	ReferenceOnly   bool
	ImportPath      string
	Model           any
	ModelIgnore     []string
	SkipValidation  bool
}

type MessageData struct {
	Name            string
	Fields          []FieldData
	Oneofs          []OneofData
	Messages        []MessageData
	ReservedNumbers []uint
	ReservedRanges  []Range
	ReservedNames   []string
	Options         []ProtoOption
	Enums           []EnumGroup
}

func (s *MessageSchema) GetFields() map[string]FieldBuilder {
	out := make(map[string]FieldBuilder)

	keys := slices.Sorted(maps.Keys(s.Fields))

	for _, k := range keys {
		f := s.Fields[k]

		data := f.GetData()
		out[data.Name] = f
	}

	return out
}

func (s *MessageSchema) GetField(n string) FieldBuilder {
	for _, f := range s.Fields {
		data := f.GetData()
		if data.Name == n {
			return f
		}
	}

	log.Fatalf("Could not find field %q in schema %q", n, s.Name)
	return nil
}

func (s *MessageSchema) CheckModel() error {
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
		err = indentErrors(fmt.Sprintf("Validation errors for model %s", modelName), err)
	}

	return err
}

func newProtoMessage(s MessageSchema, imports Set) (MessageData, error) {
	var protoFields []FieldData
	var fieldsErrors error

	if s.Model != nil && !s.SkipValidation {
		err := s.CheckModel()
		if err != nil {
			return MessageData{}, err
		}
	}

	fieldNumbers := slices.Sorted(maps.Keys(s.Fields))

	for _, fieldNr := range fieldNumbers {
		fieldBuilder := s.Fields[fieldNr]
		field, err := fieldBuilder.Build(fieldNr, imports)
		if err != nil {
			fieldsErrors = errors.Join(fieldsErrors, indentErrors(fmt.Sprintf("Errors for field %s", field.Name), err))
		} else {
			protoFields = append(protoFields, field)
		}
	}

	oneOfs := []OneofData{}
	var oneOfErrors error

	for _, oneof := range s.Oneofs {
		data, oneofErr := oneof.Build(imports)

		if oneofErr != nil {
			oneOfErrors = errors.Join(oneOfErrors, indentErrors(fmt.Sprintf("Errors for oneof %q", data.Name), oneofErr))
		}
		oneOfs = append(oneOfs, data)
	}

	subMessages := []MessageData{}
	var subMessagesErrors error

	for _, m := range s.Messages {
		data, err := newProtoMessage(m, imports)
		if err != nil {
			subMessagesErrors = errors.Join(subMessagesErrors, indentErrors(fmt.Sprintf("Errors for nested message %q", m.Name), err))
		}

		subMessages = append(subMessages, data)
	}

	if fieldsErrors != nil || oneOfErrors != nil || subMessagesErrors != nil {
		return MessageData{}, errors.Join(fieldsErrors, oneOfErrors, subMessagesErrors)
	}

	return MessageData{Name: s.Name, Fields: protoFields, ReservedNumbers: s.ReservedNumbers, ReservedRanges: s.ReservedRanges, ReservedNames: s.ReservedNames, Options: s.Options, Oneofs: oneOfs, Enums: s.Enums, Messages: subMessages}, nil
}

func MessageRef(s MessageSchema) MessageSchema {
	return MessageSchema{Name: s.Name, ReferenceOnly: true, ImportPath: s.ImportPath}
}

func Empty() MessageSchema {
	return MessageSchema{Name: "google.protobuf.Empty", ReferenceOnly: true, ImportPath: "google/protobuf/empty.proto"}
}

var (
	DisableValidator = ProtoOption{Name: "(buf.validate.message).disabled", Value: true}
	ProtoDeprecated  = ProtoOption{Name: "deprecated", Value: true}
)

func ProtoVOneof(required bool, fields ...string) ProtoOption {
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
