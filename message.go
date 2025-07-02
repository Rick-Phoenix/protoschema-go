package protoschema

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

// A map of field numbers to FieldBuilder instances.
type FieldsMap map[uint32]FieldBuilder

// A type to define a number range in a protobuf message or enum. The first number is the start, the second number is the end. For single numbers, use ReservedNumbers instead.
type Range [2]int32

// A function that receives the message data after its schema has been processed. If it returns an error, this will be gathered and logged among the other errors before exiting.
type MessageHook func(d MessageData) error

// The schema for a protobuf message. This should be created with the constructor from a FileSchema instance to automatically set the Package and File or from the constructor from another MessageSchema (if it's a nested message) to automatically set the ParentMessage field.
// It can also be used without the constructor to define custom, schema-less messages which can be used as types from fields.
type MessageSchema struct {
	// The name of the message. Use the getter to retrieve it, as it adds the parent message's prefix automatically (if there is one).
	Name string
	// The map of fields for this message. The number corresponds to the field's number in the proto file.
	Fields   FieldsMap
	oneofs   []OneofGroup
	enums    []*EnumGroup
	messages []*MessageSchema
	// The options for this message. The methods on this message and its fields that uses protovalidate rules will automatically add the necessary options to this.
	Options         []ProtoOption
	ReservedNumbers []uint
	ReservedRanges  []Range
	ReservedNames   []string
	// The struct to which this schema should conform. If nil, validation is skipped. If defined, a method will check if every field in the model (that is not included in the ModelIgnore slice) has the right name and type in the schema's output, or if there are missing or extra fields, causing a fatal error if that is the case.
	// Must be a pointer.
	Model any
	// The fields to ignore in the schema's validation. This should also be used for fields that are in the schema but not in the model, and vice versa.
	ModelIgnore []string
	// Whether to skip validation on this schema. Setting the model to nil also achieves the same behaviour.
	SkipValidation bool
	// The pointer to the FileSchema that this message belongs to. Automatically set when the message is created with the constructor from the File or Message Schema.
	File *FileSchema
	// The pointer to the ProtoFile that this message belongs to. Automatically set when the message is created with the constructor from the File or Message Schema.
	Package *ProtoPackage
	// The path to the proto file where this message is defined. Automatically set when the message is created with the constructor from the File or Message schema.
	// Use the getter to access this value as it is safer and it also handles default scenarios.
	ImportPath string
	// A function that will receive the MessageData after the schema is processed.
	Hook MessageHook
	// Custom metadata to use in the hook. Will be passed to the MessageData instance.
	Metadata map[string]any
	// If this is undefined, and the message was created with a constructor, the global ConverterFunc will be used. Otherwise, this will override it.
	ConverterFunc ConverterFunc
	// The pointer to the parent message. Use the constructor from the MessageSchema to set this automatically.
	ParentMessage *MessageSchema
}

// The output from a MessageSchema after it has been processed. This gets passed to the MessageHook, if it's defined.
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

// Gets the ImportPath field, if defined. If it's not defined, and the Package field is defined, it will fall back to a path that joins the Package's base path with this message's name in snake_case (as per the proto convention).
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

// Gets a FieldBuilder instance with a specific name, causes a fatal error if the field is not found. Modifying this field will also modify the original.
func (m *MessageSchema) GetField(n string) FieldBuilder {
	for _, f := range m.Fields {
		if f.GetName() == n {
			return f
		}
	}

	log.Fatalf("Could not find field %q in schema %q", n, m.Name)
	return nil
}

// Returns a map with the field names as keys and the FieldBuilder instances as the values. Modifying these will modify their original values.
func (m *MessageSchema) GetFields() map[string]FieldBuilder {
	out := make(map[string]FieldBuilder)

	keys := slices.Sorted(maps.Keys(m.Fields))

	for _, k := range keys {
		f := m.Fields[k]

		out[f.GetName()] = f
	}

	return out
}

// Gets the message's name, adding the prefix with the parent message's name if nested.
func (m *MessageSchema) GetName() string {
	if m == nil {
		return ""
	}

	if m.ParentMessage != nil {
		return m.ParentMessage.GetName() + "." + m.Name
	}

	return m.Name
}

// Returns the full name of the package (i.e. google.protobuf.Empty) if the package given as the argument is not the same as the MessageSchema's. Otherwise, it just returns the message's name.
func (m *MessageSchema) GetFullName(pkg *ProtoPackage) string {
	if m == nil {
		return ""
	}

	if m.Package == pkg || m.Package == nil {
		return m.GetName()
	}

	return m.Package.GetName() + "." + m.GetName()
}

// Returns true if the package given as the argument is the same as the MessageSchema's.
func (m *MessageSchema) IsInternal(p *ProtoPackage) bool {
	if m == nil || p == nil {
		return false
	}

	return m.Package == p
}

// Returns the name of the go package for this message's package, if set.
func (m *MessageSchema) GetGoPackageName() string {
	if m == nil || m.Package == nil {
		return ""
	}

	return m.Package.GetGoPackageName()
}

// Helper to generate a Cel option for this message.
func (m *MessageSchema) CelOption(id, message, expression string) {
	opt := ProtoOption{}

	opt.Name = "(buf.validate.message).cel"

	opt.Value = CelOption{Id: id, Message: message, Expression: expression}

	m.Options = append(m.Options, opt)
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
				err = errors.Join(err, fmt.Errorf("Model field %q not found in the message schema.", modelFieldName))
			}

		}
	}

	processFields(model)

	if len(msgFields) > 0 {
		for name := range msgFields {
			if !ignores.Has(name) {
				err = errors.Join(err, fmt.Errorf("Unknown field %q is not present in the message's model.", name))
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
			errAgg = errors.Join(errAgg, indentErrors(fmt.Sprintf("Errors for field %q", field.Name), err))
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

// Adds a OneofGroup to this message, automatically setting its Message, File and Package fields, while also falling back to the global OneofHook if a specific Hook is not defined.
// Returns the pointer to this OneofGroup instance.
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

// Adds a EnumGroup to this message, automatically setting its Message, File, ImportPath and Package fields.
// Returns the pointer to this EnumGroup instance.
func (m *MessageSchema) NewEnum(e EnumGroup) *EnumGroup {
	e.Message = m
	e.File = m.File
	e.Package = m.Package
	e.ImportPath = m.ImportPath
	m.enums = append(m.enums, &e)
	return &e
}

// Creates a new MessageSchema and adds it to this message's own slice of nested messages, while automatically setting the ParentMessage, File, Package and ImportPath fields.
func (m *MessageSchema) NestedMessage(nm MessageSchema) *MessageSchema {
	nm.ParentMessage = m
	nm.File = m.File
	nm.Package = m.Package
	nm.ImportPath = m.ImportPath
	m.messages = append(m.messages, &nm)
	return &nm
}
