package schemabuilder

import (
	"errors"
	"fmt"
	"log"
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

	for name, oneof := range s.Oneofs {
		data, oneofErr := oneof.Build(name, imports)

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
	Schema            *ProtoMessageSchema
	ReplaceReserved   bool
	ReplaceOptions    bool
	ReplaceOneofs     bool
	ReplaceImports    bool
	ReplaceFields     bool
	RemoveReserved    []uint
	RemoveFields      []string
	RemoveImports     []string
	RemoveOneofs      []string
	MakeReferenceOnly bool
}

func CompleteExtendProtoMessage(s ProtoMessageSchema, e ProtoMessageExtension) ProtoMessageSchema {
	newFields := make(ProtoFieldsMap)
	var hasSchema bool

	if e.Schema != nil {
		hasSchema = true
	}

	if (e.ReplaceReserved || e.ReplaceOptions || e.ReplaceOneofs || e.ReplaceImports || e.ReplaceFields) && !hasSchema {
		log.Fatalf("Tried to replace parts of the message schema for %q with a nil pointer for the replacement.", s.Name)
	}

	if hasSchema {
		maps.Copy(newFields, e.Schema.Fields)
	}

	if !e.ReplaceFields {
		maps.Copy(newFields, s.Fields)
	}

	for _, f := range e.RemoveFields {
		delete(newFields, f)
	}

	reserved := []uint{}

	if e.ReplaceReserved {
		copy(reserved, e.Schema.Reserved)
	} else {

		reserved = append(reserved, s.Reserved...)

		if hasSchema {
			reserved = append(reserved, e.Schema.Reserved...)
		}

		reserved = FilterAndDedupe(reserved, func(n uint) bool {
			return !slices.Contains(e.RemoveReserved, n)
		})
	}

	options := []ProtoOption{}

	if e.ReplaceOptions {
		copy(options, e.Schema.Options)
	} else {
		options = append(options, s.Options...)

		if hasSchema {
			options = append(options, e.Schema.Options...)
		}
	}

	if hasSchema && e.Schema.Name != "" {
		s.Name = e.Schema.Name
	}

	oneofs := make(map[string]ProtoOneOfBuilder)

	if e.ReplaceOneofs {
		maps.Copy(oneofs, e.Schema.Oneofs)
	} else {

		maps.Copy(oneofs, s.Oneofs)

		if hasSchema {
			maps.Copy(oneofs, e.Schema.Oneofs)
		}

		for _, o := range e.RemoveOneofs {
			delete(oneofs, o)
		}

		s.Oneofs = oneofs
	}

	imports := []string{}

	if e.ReplaceImports {
		copy(imports, e.Schema.Imports)
	} else {
		copy(imports, s.Imports)

		if hasSchema {
			imports = append(imports, e.Schema.Imports...)
		}

		imports = FilterAndDedupe(imports, func(i string) bool {
			return !slices.Contains(e.RemoveImports, i)
		})
	}

	s.Fields = newFields
	s.Reserved = reserved
	s.Options = options
	s.Oneofs = oneofs
	s.Imports = imports

	if e.MakeReferenceOnly {
		s.ReferenceOnly = true
	}

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
