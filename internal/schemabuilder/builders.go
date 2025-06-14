package schemabuilder

import "maps"

type ProtoService struct {
	Messages   []ProtoMessage
	FieldsFlat []string
	Imports    []string
}

type ProtoServiceSchema struct {
	Create, Get, Update, Delete *ServiceData
}

type Set map[string]struct{}

type ServiceData struct {
	Request  ProtoMessageSchema
	Response ProtoMessageSchema
}

func NewProtoService(resourceName string, s ProtoServiceSchema) ProtoService {
	out := &ProtoService{}
	imports := make(Set)
	if s.Get != nil {
		getRequest := NewProtoMessage("Get"+resourceName+"Request", s.Get.Request, imports)
		out.Messages = append(out.Messages, getRequest)
	}
	return *out
}

type ProtoFieldsMap map[string]ProtoFieldBuilder

type ProtoMessageSchema struct {
	Fields     ProtoFieldsMap
	Options    map[string]string
	CelOptions []CelFieldOpts
	Reserved   []int
}

type ProtoMessage struct {
	Name     string
	Fields   []ProtoFieldData
	Reserved []int
	Options  map[string]string
}

func NewProtoMessage(messageName string, s ProtoMessageSchema, imports Set) ProtoMessage {
	var protoFields []ProtoFieldData
	for fieldName, fieldBuilder := range s.Fields {
		protoFields = append(protoFields, fieldBuilder.Build(fieldName, imports))
	}
	return ProtoMessage{Fields: protoFields, Name: messageName, Reserved: s.Reserved, Options: s.Options}
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

// Cel field and rules aggregator, imports

type ProtoFieldBuilder interface {
	Build(fieldName string, imports Set) ProtoFieldData
}

type CelFieldOpts struct {
	Id, Message, Expression string
}

type ProtoStringBuilder struct {
	rules      map[string]string
	celOptions []CelFieldOpts
	nullable   bool
	fieldNr    int
	imports    Set
}

type MessageOption map[string]string

func ProtoString(fieldNumber int) *ProtoStringBuilder {
	return &ProtoStringBuilder{fieldNr: fieldNumber}
}

func (b *ProtoStringBuilder) Build(fieldName string, imports Set) ProtoFieldData {
	if b.nullable {
		b.imports["google/protobuf/wrappers.proto"] = struct{}{}
	}
	maps.Copy(imports, b.imports)
	return ProtoFieldData{Name: fieldName, Rules: b.rules, ColType: "string", Nullable: b.nullable, FieldNr: b.fieldNr, CelOptions: b.celOptions}
}

// Multiple can be supported so needs another method than a map
func (b *ProtoStringBuilder) CelField(o CelFieldOpts) *ProtoStringBuilder {
	b.celOptions = append(b.celOptions, CelFieldOpts{
		Id: o.Id, Expression: o.Expression, Message: o.Message,
	})

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
