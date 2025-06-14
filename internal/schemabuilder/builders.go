package schemabuilder

import (
	"log"
	"maps"
)

type ProtoServiceBuilder struct {
	messageSchemas []ProtoMessageSchema
	fieldsMap      map[string]string
}

type ProtoServiceBuilderInterface interface {
	Build() ProtoServiceOutput
}

type ProtoServiceOutput struct {
	Messages   []ProtoMessage
	FieldsFlat []string
}

type ProtoServiceSchema struct {
	Create, Get, Update, Delete *ServiceData
}

func NewProtoService(s ProtoServiceSchema) ProtoServiceOutput {
	return ProtoServiceOutput{}
}

type ProtoFields map[string]ProtoFieldBuilder

type ProtoMessageSchema struct {
	Fields  ProtoFields
	Options map[string]string
}

type ProtoMessage struct {
	Name     string
	Fields   []ProtoFieldBuilder
	Reserved []int
	Options  []string
}

type ProtoMessageBuilderInterface interface {
	Build() ProtoMessage
}

func NewProtoMessage(s ProtoMessageSchema) ProtoMessage {
	// Loop the map of fields, build them with their name as an arg and return the output
	return ProtoMessage{}
}

type ProtoField struct {
	Name    string
	Type    string
	Options map[string]string
}

type ProtoFieldData struct {
	Rules      map[string]string
	ColType    string
	Nullable   bool
	FieldNr    int
	CelOptions []CelFieldOpts
	Name       string
}

type ProtoFieldBuilder interface {
	Build(name string) ProtoFieldData
}

type CelFieldOpts struct {
	Id, Message, Expression string
}

type ProtoStringBuilder struct {
	rules      map[string]string
	celOptions []CelFieldOpts
	nullable   bool
	fieldNr    int
}

type MessageOption map[string]string

func ProtoString(fieldNumber int) *ProtoStringBuilder {
	return &ProtoStringBuilder{fieldNr: fieldNumber}
}

func (b *ProtoStringBuilder) Build(name string) ProtoFieldData {
	return ProtoFieldData{Name: name, Rules: b.rules, ColType: "string", Nullable: b.nullable, FieldNr: b.fieldNr}
}

// Multiple can be supported so needs another method than a map
func (b *ProtoStringBuilder) CelField(o CelFieldOpts) *ProtoStringBuilder {
	b.celOptions = append(b.celOptions, CelFieldOpts{
		Id: o.Id, Expression: o.Expression, Message: o.Message,
	})

	return b
}

func (b *ProtoStringBuilder) Extend(e *ProtoStringBuilder) *ProtoStringBuilder {
	extendedBuilderData := e.Build()
	if extendedBuilderData.ColType != b.Build().ColType {
		log.Fatalf("Wrong col type")
	}
	extraRules := maps.All((*e).Build().Rules)
	maps.Insert(b.rules, extraRules)
	return b
}

func (b *ProtoStringBuilder) Nullable() *ProtoStringBuilder {
	b.nullable = true
	return b
}

func (b *ProtoStringBuilder) Required() *ProtoStringBuilder {
	b.rules["(buf.validate.field).required"] = "true"
	return b
}

type Int64ColumnBuilder struct {
	rules    map[string]string
	nullable bool
	fieldNr  int
}

func Int64Col() *Int64ColumnBuilder {
	return &Int64ColumnBuilder{}
}

func (b *Int64ColumnBuilder) Nullable() *Int64ColumnBuilder {
	b.nullable = true
	return b
}

func (b *Int64ColumnBuilder) Build() ProtoFieldData {
	return ProtoFieldData{Rules: b.rules, ColType: "int64", Nullable: b.nullable, FieldNr: b.fieldNr}
}

type FieldMaskBuilder struct {
	fieldNr int
}

func FieldMask(fieldNumber int) *FieldMaskBuilder {
	return &FieldMaskBuilder{fieldNr: fieldNumber}
}

func (b *FieldMaskBuilder) Build() ProtoFieldData {
	return ProtoFieldData{FieldNr: b.fieldNr, ColType: "fieldMask"}
}

// type BytesColumnBuilder struct {
// 	rules     []string
// 	requests  []string
// 	responses []string
// 	nullable  bool
// }
//
// // The generic type parameter is a slice of bytes
// func (b *BytesColumnBuilder) Build() Column {
// 	return Column{Rules: b.rules, Requests: b.requests, Responses: b.responses, ColType: "byte[]", Nullable: b.nullable}
// }
//
// func BytesCol() *BytesColumnBuilder {
// 	return &BytesColumnBuilder{}
// }
//
// func (b *BytesColumnBuilder) Nullable() *BytesColumnBuilder {
// 	b.nullable = true
// 	return b
// }
//
// func (b *BytesColumnBuilder) Requests(r ...string) *BytesColumnBuilder {
// 	b.requests = append(b.requests, r...)
// 	return b
// }
//
// func (b *BytesColumnBuilder) Responses(r ...string) *BytesColumnBuilder {
// 	b.responses = append(b.responses, r...)
// 	return b
// }
//
// type TimeStampColumnBuilder struct {
// 	rules     []string
// 	requests  []string
// 	responses []string
// }
//
// func TimestampCol() *TimeStampColumnBuilder {
// 	return &TimeStampColumnBuilder{}
// }
//
// func (b *TimeStampColumnBuilder) Build() Column {
// 	return Column{Rules: b.rules, Requests: b.requests, Responses: b.responses, ColType: "timestamp"}
// }
//
// func (b *TimeStampColumnBuilder) Requests(r ...string) *TimeStampColumnBuilder {
// 	b.requests = append(b.requests, r...)
// 	return b
// }
//
// func (b *TimeStampColumnBuilder) Responses(r ...string) *TimeStampColumnBuilder {
// 	b.responses = append(b.responses, r...)
// 	return b
// }
