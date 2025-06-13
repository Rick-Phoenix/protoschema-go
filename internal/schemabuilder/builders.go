package schemabuilder

import (
	"log"
	"maps"
)

type ColumnsMap map[string]ColumnBuilder

type TableBuilder struct {
	Name    string
	Columns ColumnsMap
}

type ServiceData struct {
	Request  ColumnBuilder
	Response ColumnBuilder
}

type ServiceOutput struct {
	Request  Column
	Response Column
}

type MethodsData struct {
	Create, Get, Update, Delete *ServiceData
}

type MethodsOut struct {
	Create, Get, Update, Delete *ServiceOutput
}

// Think about how to implement CEL rules
// Extend with other column builder
// Rules as map to avoid duplicates
// Inherit type directly from schema (but explicitly needed for correct rules still)
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
}

func StringCol(m *MethodsData) *MethodsOut {
	out := &MethodsOut{}
	out.Get.Request = m.Get.Request.Build()
	out.Get.Response = m.Get.Response.Build()
	out.Create.Request = m.Create.Request.Build()
	out.Create.Response = m.Create.Response.Build()
	out.Update.Request = m.Update.Request.Build()
	out.Update.Response = m.Update.Response.Build()
	out.Delete.Request = m.Delete.Request.Build()
	out.Delete.Response = m.Delete.Response.Build()

	return out
}

func StrValid() *StringColumnBuilder {

	return &StringColumnBuilder{}
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
	return Column{Rules: b.rules, ColType: "string", Nullable: b.nullable}
}

type Int64ColumnBuilder struct {
	rules    map[string]string
	nullable bool
}

func Int64Col() *Int64ColumnBuilder {
	return &Int64ColumnBuilder{}
}

func (b *Int64ColumnBuilder) Nullable() *Int64ColumnBuilder {
	b.nullable = true
	return b
}

func (b *Int64ColumnBuilder) Build() Column {
	return Column{Rules: b.rules, ColType: "int64", Nullable: b.nullable}
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
