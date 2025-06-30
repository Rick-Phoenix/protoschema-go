package schemabuilder

import (
	"errors"
	"fmt"
	"log"
	"maps"
	"reflect"
	"slices"

	u "github.com/Rick-Phoenix/goutils"
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
	Model           any
	ModelIgnore     []string
	SkipValidation  bool
	File            *FileSchema
	Package         *ProtoPackage
	ImportPath      string
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
	File            *FileSchema
	Package         *ProtoPackage
}

type modelField struct {
	Name       string
	IsInternal bool
}

type messageConverter struct {
	TimestampFields Set
	Resource        string
	SrcType         string
	Fields          []modelField
}

type ConverterData struct {
	Package            string
	GoPackage          string
	Imports            Set
	MessageConverters  []*messageConverter
	RepeatedConverters Set
}

func (m *MessageSchema) GetFields() map[string]FieldBuilder {
	out := make(map[string]FieldBuilder)

	keys := slices.Sorted(maps.Keys(m.Fields))

	for _, k := range keys {
		f := m.Fields[k]

		data := f.GetData()
		out[data.Name] = f
	}

	return out
}

func (m MessageSchema) GetFullName(pkg *ProtoPackage) string {
	if m.Package == nil {
		return ""
	}

	if m.Package == pkg {
		return m.Name
	}

	return m.Package.Name + "." + m.Name
}

func (m *MessageSchema) IsInternal(p *ProtoPackage) bool {
	if m == nil || p == nil {
		return false
	}

	return m.Package == p
}

func (m *MessageSchema) GetField(n string) FieldBuilder {
	for _, f := range m.Fields {
		data := f.GetData()
		if data.Name == n {
			return f
		}
	}

	log.Fatalf("Could not find field %q in schema %q", n, m.Name)
	return nil
}

type ConverterFuncData struct {
	Package    *ProtoPackage
	File       *FileSchema
	Message    *MessageSchema
	ModelField reflect.StructField
	ProtoField FieldBuilder
}

type ConverterFunc func(ConverterFuncData)

func (m *MessageSchema) CreateFieldConverter(converter *messageConverter, field reflect.StructField, pfield FieldBuilder) {
	fieldConvData := modelField{Name: field.Name}
	isTime := field.Type.String() == "time.Time"

	if isTime {
		converter.TimestampFields[field.Name] = present
		m.Package.Converter.Imports["google.golang.org/protobuf/types/known/timestamppb"] = present
	}

	if pfield.GetData().IsNonScalar && !isTime {
		m.Package.Converter.Imports[getPkgPath(field.Type)] = present

		if msgRef := pfield.GetMessageRef(); msgRef != nil && msgRef.Model != nil {
			if msgRef.IsInternal(m.Package) {
				fieldConvData.IsInternal = true
				if pfield.IsRepeated() {
					m.Package.Converter.RepeatedConverters[msgRef.Name] = present
				}
			}
		}
	}

	converter.Fields = append(converter.Fields, fieldConvData)
}

func (m MessageSchema) CheckModel() error {
	model := reflect.TypeOf(m.Model).Elem()
	modelName := model.String()
	msgFields := m.GetFields()
	ignores := u.NewSet(m.ModelIgnore...)

	conv := &messageConverter{
		Resource:        m.Name,
		SrcType:         modelName,
		TimestampFields: make(Set),
	}

	hasConverterFunc := m.Package != nil && m.Package.converterFunc != nil

	if !hasConverterFunc {
		m.Package.Converter.MessageConverters = append(m.Package.Converter.MessageConverters, conv)
		m.Package.Converter.Imports[getPkgPath(model)] = present
	}

	var err error

	var processFields func(t reflect.Type)
	processFields = func(t reflect.Type) {
		for i := range t.NumField() {
			field := t.Field(i)
			if field.Anonymous {
				embeddedType := field.Type
				if embeddedType.Kind() == reflect.Struct {
					// Recursive
					processFields(embeddedType)
					continue
				}
			}
			modelFieldName := field.Tag.Get("json")
			if modelFieldName == "" {
				modelFieldName = toSnakeCase(field.Name)
			}
			ignore := ignores.Has(modelFieldName)
			fieldType := field.Type.String()

			if pfield, exists := msgFields[modelFieldName]; exists {
				if hasConverterFunc {
					m.Package.converterFunc(ConverterFuncData{
						Package: m.Package, File: m.File, Message: &m, ModelField: field, ProtoField: pfield,
					})
				} else {
					m.CreateFieldConverter(conv, field, pfield)
				}

				delete(msgFields, modelFieldName)

				if ignore {
					continue
				}

				goType := pfield.GetGoType()
				fieldName := pfield.GetName()

				if pfield.GetGoType() != fieldType && !ignores.Has(fieldName) {
					err = errors.Join(err, fmt.Errorf("Expected type %q for field %q, found %q.", fieldType, modelFieldName, goType))
				}
			} else if !ignore {
				err = errors.Join(err, fmt.Errorf("Column %q not found in the proto schema for %q.", modelFieldName, t))
			}

		}
	}

	processFields(model)

	if len(msgFields) > 0 {
		for name := range msgFields {
			if !ignores.Has(name) {
				err = errors.Join(err, fmt.Errorf("Unknown field %q found in the message schema for model %q.", name, modelName))
			}
		}
	}

	if err != nil {
		err = indentErrors(fmt.Sprintf("Validation errors for model %s", modelName), err)
	}

	return err
}

func (m *MessageSchema) Build(imports Set) (MessageData, error) {
	var protoFields []FieldData
	var fieldsErrors error

	if m.Model != nil && !m.SkipValidation {
		err := m.CheckModel()
		if err != nil {
			return MessageData{}, err
		}
	}

	fieldNumbers := slices.Sorted(maps.Keys(m.Fields))

	for _, fieldNr := range fieldNumbers {
		fieldBuilder := m.Fields[fieldNr]
		field, err := fieldBuilder.Build(fieldNr, imports)
		if err != nil {
			fieldsErrors = errors.Join(fieldsErrors, indentErrors(fmt.Sprintf("Errors for field %s", field.Name), err))
		} else {
			protoFields = append(protoFields, field)
		}
	}

	oneOfs := []OneofData{}
	var oneOfErrors error

	for _, oneof := range m.Oneofs {
		data, oneofErr := oneof.Build(imports)

		if oneofErr != nil {
			oneOfErrors = errors.Join(oneOfErrors, indentErrors(fmt.Sprintf("Errors for oneof %q", data.Name), oneofErr))
		}
		oneOfs = append(oneOfs, data)
	}

	subMessages := []MessageData{}
	var subMessagesErrors error

	for _, m := range m.Messages {
		data, err := m.Build(imports)
		if err != nil {
			subMessagesErrors = errors.Join(subMessagesErrors, indentErrors(fmt.Sprintf("Errors for nested message %q", m.Name), err))
		}

		subMessages = append(subMessages, data)
	}

	if fieldsErrors != nil || oneOfErrors != nil || subMessagesErrors != nil {
		return MessageData{}, errors.Join(fieldsErrors, oneOfErrors, subMessagesErrors)
	}

	return MessageData{Name: m.Name, Fields: protoFields, ReservedNumbers: m.ReservedNumbers, ReservedRanges: m.ReservedRanges, ReservedNames: m.ReservedNames, Options: m.Options, Oneofs: oneOfs, Enums: m.Enums, Messages: subMessages, File: m.File, Package: m.Package}, nil
}

func Empty() *MessageSchema {
	return &MessageSchema{Name: "Empty", ImportPath: "google/protobuf/empty.proto", Package: &ProtoPackage{Name: "google.protobuf", goPackageName: "emptypb", goPackagePath: "google.golang.org/protobuf/types/known/emptypb"}}
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
