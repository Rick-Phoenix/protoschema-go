package schemabuilder

import "reflect"

type ConverterFunc func(ConverterFuncData)

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
