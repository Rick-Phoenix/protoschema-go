package schemabuilder

import (
	"errors"
	"fmt"
	"log"
	"maps"
	"path"
	"reflect"
	"slices"

	u "github.com/Rick-Phoenix/goutils"
)

type FieldsMap map[uint32]FieldBuilder

type Range [2]int32

type MessageHook func(d MessageData) error

type MessageSchema struct {
	Name            string
	Fields          FieldsMap
	oneofs          []OneofGroup
	enums           []*EnumGroup
	messages        []*MessageSchema
	Options         []ProtoOption
	ReservedNumbers []uint
	ReservedRanges  []Range
	ReservedNames   []string
	Model           any
	ModelIgnore     []string
	SkipValidation  bool
	File            *FileSchema
	Package         *ProtoPackage
	ImportPath      string
	Hook            MessageHook
	Metadata        map[string]any
	ConverterFunc   ConverterFunc
	ParentMessage   *MessageSchema
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
	Metadata        map[string]any
}

func (m *MessageSchema) GetImportPath() string {
	if m == nil {
		return ""
	}

	if m.ImportPath == "" {
		if filePath := m.File.GetImportPath(); filePath != "" {
			return filePath
		}

		if pkgBasePath := m.Package.GetBasePath(); pkgBasePath != "" {
			return path.Join(pkgBasePath, addMissingSuffix(toSnakeCase(m.Name), ".proto"))
		}
	}

	return m.ImportPath
}

func (m *MessageSchema) GetFields() map[string]FieldBuilder {
	out := make(map[string]FieldBuilder)

	keys := slices.Sorted(maps.Keys(m.Fields))

	for _, k := range keys {
		f := m.Fields[k]

		out[f.GetName()] = f
	}

	return out
}

func (m *MessageSchema) GetName() string {
	if m == nil {
		return ""
	}

	if m.ParentMessage != nil {
		return m.ParentMessage.GetName() + "." + m.Name
	}

	return m.Name
}

func (m *MessageSchema) GetFullName(pkg *ProtoPackage) string {
	if m == nil {
		return ""
	}

	if m.Package == pkg || m.Package == nil {
		return m.GetName()
	}

	return m.Package.GetName() + "." + m.GetName()
}

func (m *MessageSchema) IsInternal(p *ProtoPackage) bool {
	if m == nil || p == nil {
		return false
	}

	return m.Package == p
}

func (m *MessageSchema) GetField(n string) FieldBuilder {
	for _, f := range m.Fields {
		if f.GetName() == n {
			return f
		}
	}

	log.Fatalf("Could not find field %q in schema %q", n, m.Name)
	return nil
}

func (m *MessageSchema) GetGoPackageName() string {
	if m == nil || m.Package == nil {
		return ""
	}

	return m.Package.GetGoPackageName()
}

func (m *MessageSchema) checkModel() error {
	model := reflect.TypeOf(m.Model).Elem()
	modelName := model.String()
	msgFields := m.GetFields()
	ignores := u.NewSet(m.ModelIgnore...)

	conv := &messageConverter{
		Resource:        m.Name,
		SrcType:         modelName,
		TimestampFields: make(Set),
	}

	hasConverterFunc := m.Package != nil && m.ConverterFunc != nil

	if !hasConverterFunc {
		m.Package.converter.MessageConverters = append(m.Package.converter.MessageConverters, conv)
		m.Package.converter.Imports[getPkgPath(model)] = present
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
				}
				continue
			}
			modelFieldName := field.Tag.Get("json")
			if modelFieldName == "" {
				modelFieldName = toSnakeCase(field.Name)
			}
			ignore := ignores.Has(modelFieldName)
			fieldType := field.Type.String()

			if pfield, exists := msgFields[modelFieldName]; exists {
				if hasConverterFunc {
					m.ConverterFunc(ConverterFuncData{
						Package: m.Package, File: m.File, Message: m, ModelField: field, ProtoField: pfield,
					})
				} else {
					m.createFieldConverter(conv, field, pfield)
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

func (m *MessageSchema) build(imports Set) (MessageData, error) {
	var protoFields []FieldData
	var errAgg error

	if m.Model != nil && !m.SkipValidation {
		err := m.checkModel()
		if err != nil {
			return MessageData{}, err
		}
	}

	fieldNumbers := slices.Sorted(maps.Keys(m.Fields))

	for _, fieldNr := range fieldNumbers {
		fieldBuilder := m.Fields[fieldNr]
		field, err := fieldBuilder.Build(fieldNr, imports)
		if err != nil {
			errAgg = errors.Join(errAgg, indentErrors(fmt.Sprintf("Errors for field %s", field.Name), err))
		} else {
			protoFields = append(protoFields, field)
		}
	}

	oneOfs := []OneofData{}

	for _, oneof := range m.oneofs {
		data, oneofErr := oneof.build(imports)

		if oneofErr != nil {
			errAgg = errors.Join(errAgg, indentErrors(fmt.Sprintf("Errors for oneof %q", data.Name), oneofErr))
		}
		oneOfs = append(oneOfs, data)
	}

	subMessages := []MessageData{}

	for _, m := range m.messages {
		data, err := m.build(imports)
		if err != nil {
			errAgg = errors.Join(errAgg, indentErrors(fmt.Sprintf("Errors for nested message %q", m.Name), err))
		}

		subMessages = append(subMessages, data)
	}

	out := MessageData{Name: m.Name, Fields: protoFields, ReservedNumbers: m.ReservedNumbers, ReservedRanges: m.ReservedRanges, ReservedNames: m.ReservedNames, Options: m.Options, Oneofs: oneOfs, Enums: u.ToValSlice(m.enums), Messages: subMessages, File: m.File, Package: m.Package, Metadata: m.Metadata}

	if m.Hook != nil {
		err := m.Hook(out)
		if err != nil {
			errAgg = errors.Join(errAgg, indentErrors("Error in message hook", err))
		}
	}

	return out, errAgg
}

func (m *MessageSchema) NewOneof(of OneofGroup) *OneofGroup {
	of.Message = m
	of.File = m.File
	of.Package = m.Package
	m.oneofs = append(m.oneofs, of)
	if of.Hook == nil && m.Package != nil {
		of.Hook = m.Package.oneofHook
	}
	return &of
}

func (m *MessageSchema) NewEnum(e EnumGroup) *EnumGroup {
	e.Message = m
	e.File = m.File
	e.Package = m.Package
	e.ImportPath = m.ImportPath
	m.enums = append(m.enums, &e)
	return &e
}

func (m *MessageSchema) NestedMessage(nm MessageSchema) *MessageSchema {
	nm.ParentMessage = m
	nm.File = m.File
	nm.Package = m.Package
	nm.ImportPath = m.ImportPath
	m.messages = append(m.messages, &nm)
	return &nm
}
