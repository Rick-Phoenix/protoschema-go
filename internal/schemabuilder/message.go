package schemabuilder

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

type ProtoFieldsMap map[string]ProtoFieldBuilder

type ProtoMessageSchema struct {
	Name          string
	Fields        ProtoFieldsMap
	Oneofs        map[string]ProtoOneOfBuilder
	Options       []ProtoOption
	Reserved      []uint
	ReferenceOnly bool
	Imports       []string
}

type ProtoMessage struct {
	Name     string
	Fields   []ProtoFieldData
	Oneofs   []ProtoOneOfData
	Reserved []uint
	Options  []ProtoOption
}

// Using pointers to make these more composable
// Not very extensible/reusable in general, it's better to make them composable
// Except for reference only messages
// Should also be easier to override field number for reusable fields
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

	// adapt this for map
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

	return ProtoMessage{Name: s.Name, Fields: protoFields, Reserved: s.Reserved, Options: s.Options, Oneofs: oneOfs}, nil
}

func ProtoMessageReference(name string, imports ...string) ProtoMessageSchema {
	return ProtoMessageSchema{Name: name, ReferenceOnly: true, Imports: imports}
}

func ProtoEmpty() ProtoMessageSchema {
	return ProtoMessageSchema{Name: "google.protobuf.Empty", ReferenceOnly: true, Imports: []string{"google/protobuf/empty.proto"}}
}

type ProtoMessageExtension struct {
	Schema          *ProtoMessageSchema
	ReferenceOnly   bool
	ReplaceReserved bool
	ReplaceOptions  bool
	ReplaceOneofs   bool
	ReplaceImports  bool
	RemoveReserved  []uint
	RemoveFields    []string
	RemoveImports   []string
	RemoveOneofs    []string
}

func CompleteExtendProtoMessage(s ProtoMessageSchema, e ProtoMessageExtension) ProtoMessageSchema {

	newFields := make(ProtoFieldsMap)
	MapsMultiCopy(newFields, s.Fields, e.Schema.Fields)

	for _, f := range e.RemoveFields {
		delete(newFields, f)
	}

	reserved := []uint{}

	// Check for nil
	if e.ReplaceReserved {
		reserved = e.Schema.Reserved
	} else {

		reserved = slices.Concat(s.Reserved, e.Schema.Reserved)

		reserved = FilterAndDedupe(reserved, func(n uint) bool {
			return !slices.Contains(e.RemoveReserved, n)
		})
	}

	options := []ProtoOption{}

	if e.ReplaceOptions {
		options = e.Schema.Options
	} else {
		options = slices.Concat(s.Options, e.Schema.Options)
	}

	if e.Schema.Name != "" {
		s.Name = e.Schema.Name
	}

	if e.ReplaceOneofs {
		s.Oneofs = e.Schema.Oneofs
	} else {
		new := make(map[string]ProtoOneOfBuilder)

		MapsMultiCopy(new, s.Oneofs, e.Schema.Oneofs)

		for _, o := range e.RemoveOneofs {
			delete(new, o)
		}

		s.Oneofs = new
	}

	if e.ReplaceImports {
		s.Imports = e.Schema.Imports
	} else {
		s.Imports = slices.Concat(s.Imports, e.Schema.Imports)
		s.Imports = FilterAndDedupe(s.Imports, func(i string) bool {
			return !slices.Contains(e.RemoveImports, i)
		})
	}

	s.Fields = newFields
	s.Reserved = reserved
	s.Options = options
	s.ReferenceOnly = e.ReferenceOnly

	return s
}

func ExtendProtoMessage(s ProtoMessageSchema, ext *ProtoMessageSchema) ProtoMessageSchema {
	if ext == nil {
		return s
	}

	newFields := make(ProtoFieldsMap)
	maps.Copy(newFields, s.Fields)
	maps.Copy(newFields, ext.Fields)

	reserved := slices.Concat(s.Reserved, ext.Reserved)
	reserved = Dedupe(reserved)

	s.Fields = newFields
	s.Reserved = reserved
	s.Options = ext.Options
	s.Name = ext.Name

	return s
}

func OmitProtoMessage(s ProtoMessageSchema, keys ...string) *ProtoMessageSchema {
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

func MessageCelOption(o CelFieldOpts) ProtoOption {
	out := ProtoOption{}

	out.Name = "(buf.validate.message).cel"

	out.Value = GetCelOption(o)

	return out
}
