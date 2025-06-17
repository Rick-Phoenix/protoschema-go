package schemabuilder

import (
	"maps"
	"slices"
)

type ProtoFieldsMap map[string]ProtoFieldBuilder

type MessageOption struct {
	Name  string
	Value string
}

type ProtoMessageSchema struct {
	Fields     ProtoFieldsMap
	OneOfs     []ProtoOneOfSchema
	Options    []MessageOption
	CelOptions []CelFieldOpts
	Reserved   []int
}

type ProtoMessage struct {
	Fields     []ProtoFieldData
	OneOfs     []ProtoOneOfData
	Reserved   []int
	CelOptions []CelFieldOpts
	Options    []MessageOption
}

func NewProtoMessage(s ProtoMessageSchema, imports Set) (ProtoMessage, Errors) {
	var protoFields []ProtoFieldData
	var errors Errors

	for fieldName, fieldBuilder := range s.Fields {
		field, err := fieldBuilder.Build(fieldName, imports)
		if err != nil {
			errors = append(errors, err)
		} else {
			protoFields = append(protoFields, field)
		}
	}

	oneOfs := []ProtoOneOfData{}

	for _, oneof := range s.OneOfs {
		data := ProtoOneOfData{}
		data.Name = oneof.Name
		data.Options = oneof.Options

		for name, field := range oneof.Choices {
			oneOfField, err := field.Build(name, imports)
			if err != nil {
				errors = append(errors, err)
			} else {
				data.Choices = append(data.Choices, oneOfField)
			}
		}

		oneOfs = append(oneOfs, data)
	}

	if len(errors) > 0 {
		return ProtoMessage{}, errors
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
