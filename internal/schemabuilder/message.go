package schemabuilder

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

type ProtoFieldsMap map[string]ProtoFieldBuilder

type ProtoMessageSchema struct {
	Fields     ProtoFieldsMap
	OneOfs     []ProtoOneOfBuilder
	Options    []ProtoOption
	CelOptions []CelFieldOpts
	Reserved   []int
}

type ProtoMessage struct {
	Fields     []ProtoFieldData
	OneOfs     []ProtoOneOfData
	Reserved   []int
	CelOptions []CelFieldOpts
	Options    []ProtoOption
}

func NewProtoMessage(s ProtoMessageSchema, imports Set) (ProtoMessage, error) {
	var protoFields []ProtoFieldData
	var fieldsErrors error

	for fieldName, fieldBuilder := range s.Fields {
		field, err := fieldBuilder.Build(fieldName, imports)
		if err != nil {
			fieldsErrors = errors.Join(fieldsErrors, IndentErrors(fmt.Sprintf("Errors for field %s", fieldName), err))
		} else {
			protoFields = append(protoFields, field)
		}
	}

	oneOfs := []ProtoOneOfData{}
	var oneOfErrors error

	for _, oneof := range s.OneOfs {
		data, oneofErr := oneof.Build(imports)

		if oneofErr != nil {
			oneOfErrors = errors.Join(oneOfErrors, IndentErrors(fmt.Sprintf("Errors for oneOf member %s", data.Name), oneofErr))
		}
		oneOfs = append(oneOfs, data)
	}

	if fieldsErrors != nil || oneOfErrors != nil {
		return ProtoMessage{}, errors.Join(fieldsErrors, oneOfErrors)
	}

	return ProtoMessage{Fields: protoFields, Reserved: s.Reserved, Options: s.Options, CelOptions: s.CelOptions, OneOfs: oneOfs}, nil
}

func ExtendProtoMessage(s ProtoMessageSchema, override *ProtoMessageSchema) *ProtoMessageSchema {
	if override == nil {
		return &s
	}
	newFields := make(ProtoFieldsMap)
	maps.Copy(newFields, s.Fields)
	maps.Copy(newFields, override.Fields)

	newCelOptions := slices.Concat(s.CelOptions, override.CelOptions)
	newCelOptions = DedupeNonComp(newCelOptions)

	reserved := slices.Concat(s.Reserved, override.Reserved)
	reserved = Dedupe(reserved)

	s.Fields = newFields
	s.Reserved = reserved
	s.Options = override.Options
	s.CelOptions = newCelOptions

	return &s
}

func OmitProtoMessage(s ProtoMessageSchema, keys []string) *ProtoMessageSchema {
	newFields := make(ProtoFieldsMap)
	maps.Copy(newFields, s.Fields)

	for _, key := range keys {
		delete(newFields, key)
	}

	s.Fields = newFields

	return &s
}

var DisableValidator = ProtoOption{Name: "(buf.validate.message).disabled", Value: "true"}
var ProtoDeprecated = ProtoOption{Name: "deprecated", Value: "true"}

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
