package schemabuilder

import "reflect"

// By default, this package will try to automatically generate functions that can convert messages with specific models (like database items) into their respective message type. If this function is defined, it will take over that role, and it will receive the data for each message field.
type ConverterFunc func(ConverterFuncData)

// The data passed to the ConverterFunc, if defined. When iterating the fields of a message schema's model, this function will be called with the reflect.Structfield value for that struct field, along with the respective FieldBuilder instance and the surrounding Package, File and Message instances.
type ConverterFuncData struct {
	Package    *ProtoPackage
	File       *FileSchema
	Message    *MessageSchema
	ModelField reflect.StructField
	ProtoField FieldBuilder
}

type modelFieldData struct {
	Name       string
	IsInternal bool
}

type messageConverter struct {
	TimestampFields Set
	Resource        string
	SrcType         string
	Fields          []modelFieldData
}

type converterData struct {
	Package            string
	GoPackage          string
	Imports            Set
	MessageConverters  []*messageConverter
	RepeatedConverters Set
}

func (m *MessageSchema) createFieldConverter(converter *messageConverter, modelField reflect.StructField, pfield FieldBuilder) {
	fieldConvData := modelFieldData{Name: modelField.Name}
	isTime := modelField.Type.String() == "time.Time"

	if isTime {
		converter.TimestampFields[modelField.Name] = present
		m.Package.converter.Imports["google.golang.org/protobuf/types/known/timestamppb"] = present
	}

	if pfield.IsNonScalar() && !isTime {
		m.Package.converter.Imports[getPkgPath(modelField.Type)] = present

		if msgRef := pfield.GetMessageRef(); msgRef != nil && msgRef.Model != nil {
			if msgRef.IsInternal(m.Package) {
				fieldConvData.IsInternal = true
				if pfield.IsRepeated() {
					m.Package.converter.RepeatedConverters[msgRef.Name] = present
				}
			}
		}
	}

	converter.Fields = append(converter.Fields, fieldConvData)
}
