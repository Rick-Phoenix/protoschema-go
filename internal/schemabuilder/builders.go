package schemabuilder

import (
	"log"
	"maps"
)

type ColumnsMap map[string]ColumnBuilder

// Think about how to implement CEL rules
// FieldMask method
type Column struct {
	Rules    map[string]string
	ColType  string
	Nullable bool
	FieldNr  int
}

type ColumnBuilder interface {
	Build() Column
}

type StringColumnBuilder struct {
	rules    map[string]string
	nullable bool
	fieldNr  int
}

func ProtoString(fieldNumber int) *StringColumnBuilder {

	return &StringColumnBuilder{fieldNr: fieldNumber}
}

func (b *StringColumnBuilder) Extend(e *StringColumnBuilder) *StringColumnBuilder {
	extendedBuilderData := e.Build()
	if extendedBuilderData.ColType != b.Build().ColType {
		log.Fatalf("Wrong col type")
	}
	extraRules := maps.All((*e).Build().Rules)
	maps.Insert(b.rules, extraRules)
	return b
}

func (b *StringColumnBuilder) Nullable() *StringColumnBuilder {
	b.nullable = true
	return b
}

func (b *StringColumnBuilder) Required() *StringColumnBuilder {
	b.rules["(buf.validate.field).required"] = "true"
	return b
}

func (b *StringColumnBuilder) Build() Column {
	return Column{Rules: b.rules, ColType: "string", Nullable: b.nullable, FieldNr: b.fieldNr}
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

func (b *Int64ColumnBuilder) Build() Column {
	return Column{Rules: b.rules, ColType: "int64", Nullable: b.nullable, FieldNr: b.fieldNr}
}

type FieldMaskBuilder struct {
	fieldNr int
}

func FieldMask(fieldNumber int) *FieldMaskBuilder {
	return &FieldMaskBuilder{fieldNr: fieldNumber}
}

func (b *FieldMaskBuilder) Build() Column {
	return Column{FieldNr: b.fieldNr, ColType: "fieldMask"}
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
