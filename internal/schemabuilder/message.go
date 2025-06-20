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
	ImportPath    string
}

type ProtoMessage struct {
	Name     string
	Fields   []ProtoFieldData
	Oneofs   []ProtoOneOfData
	Reserved []uint
	Options  []ProtoOption
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

func ImportedMessage(name string, importPath string) ProtoMessageSchema {
	return ProtoMessageSchema{Name: name, ReferenceOnly: true, ImportPath: importPath}
}

func MessageRef(name string) ProtoMessageSchema {
	return ProtoMessageSchema{Name: name, ReferenceOnly: true}
}

func ProtoEmpty() ProtoMessageSchema {
	return ProtoMessageSchema{Name: "google.protobuf.Empty", ReferenceOnly: true, ImportPath: "google/protobuf/empty.proto"}
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

func ExtendProtoMessage(s *ProtoMessageSchema, e ProtoMessageExtension) ProtoMessageSchema {
	if s == nil {
		log.Fatalf("Received a nil pointer when trying to extend a message schema.")
	}

	var hasSchema bool
	if e.Schema != nil {
		hasSchema = true
	}

	if (e.ReplaceReserved || e.ReplaceOptions || e.ReplaceOneofs || e.ReplaceImports || e.ReplaceFields) && !hasSchema {
		log.Fatalf("Tried to replace parts of the message schema for %q with a nil pointer for the replacement.", s.Name)
	}

	newSchema := ProtoMessageSchema{}

	newFields := make(ProtoFieldsMap)

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

	}

	newSchema.Fields = newFields
	newSchema.Reserved = reserved
	newSchema.Options = options
	newSchema.Oneofs = oneofs

	if hasSchema && e.Schema.Name != "" {
		newSchema.Name = e.Schema.Name
	}

	if e.MakeReferenceOnly {
		newSchema.ReferenceOnly = true
	}

	return newSchema
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
