package schemabuilder

import (
	"fmt"
)

type ColumnsMap map[string]ColumnBuilder

type TableBuilder struct {
	Name    string
	Columns ColumnsMap
}

var UserSchema = &TableBuilder{
	Name: "User",
	Columns: ColumnsMap{
		"Name":      StringCol().Required().MinLen(3).Requests("create").Responses("get", "create"),
		"Age":       Int64Col().Responses("get"),
		"Blob":      BytesCol().Requests("get"),
		"CreatedAt": TimestampCol().Responses("get"),
	},
}

type Column struct {
	Rules     []string
	Requests  []string
	Responses []string
	ColType   string
}

type ColumnBuilder interface {
	Build() Column
}

type StringColumnBuilder struct {
	rules     []string
	requests  []string
	responses []string
}

func StringCol() *StringColumnBuilder {
	return &StringColumnBuilder{}
}

func (b *StringColumnBuilder) Requests(r ...string) *StringColumnBuilder {
	b.requests = append(b.requests, r...)
	return b
}

func (b *StringColumnBuilder) Responses(r ...string) *StringColumnBuilder {
	b.responses = append(b.responses, r...)
	return b
}

func (b *StringColumnBuilder) Required() *StringColumnBuilder {
	b.rules = append(b.rules, "(buf.validate.field).required = true")
	return b
}

func (b *StringColumnBuilder) MinLen(len uint64) *StringColumnBuilder {
	rule := fmt.Sprintf("(buf.validate.field).string.min_len = %d", len)
	b.rules = append(b.rules, rule)
	return b
}

func (b *StringColumnBuilder) Len(len uint64) *StringColumnBuilder {
	rule := fmt.Sprintf("(buf.validate.field).string.len = %d", len)
	b.rules = append(b.rules, rule)
	return b
}

func (b *StringColumnBuilder) MaxLen(len uint64) *StringColumnBuilder {
	rule := fmt.Sprintf("(buf.validate.field).string.max_len = %d", len)
	b.rules = append(b.rules, rule)
	return b
}

func (b *StringColumnBuilder) LenBytes(len uint64) *StringColumnBuilder {
	rule := fmt.Sprintf("(buf.validate.field).string.len_bytes = %d", len)
	b.rules = append(b.rules, rule)
	return b
}

func (b *StringColumnBuilder) MinBytes(len uint64) *StringColumnBuilder {
	rule := fmt.Sprintf("(buf.validate.field).string.min_bytes = %d", len)
	b.rules = append(b.rules, rule)
	return b
}

func (b *StringColumnBuilder) MaxBytes(len uint64) *StringColumnBuilder {
	rule := fmt.Sprintf("(buf.validate.field).string.max_bytes = %d", len)
	b.rules = append(b.rules, rule)
	return b
}

func (b *StringColumnBuilder) Email() *StringColumnBuilder {
	rule := fmt.Sprintf("(buf.validate.field).string.email = true")
	b.rules = append(b.rules, rule)
	return b
}

func (b *StringColumnBuilder) Build() Column {
	return Column{Rules: b.rules, Requests: b.requests, Responses: b.responses, ColType: "string"}
}

type Int64ColumnBuilder struct {
	rules     []string
	requests  []string
	responses []string
}

func Int64Col() *Int64ColumnBuilder {
	return &Int64ColumnBuilder{}
}

func (b *Int64ColumnBuilder) Requests(r ...string) *Int64ColumnBuilder {
	b.requests = append(b.requests, r...)
	return b
}

func (b *Int64ColumnBuilder) Responses(r ...string) *Int64ColumnBuilder {
	b.responses = append(b.responses, r...)
	return b
}

func (b *Int64ColumnBuilder) Build() Column {
	return Column{Rules: b.rules, Requests: b.requests, Responses: b.responses, ColType: "int64"}
}

type BytesColumnBuilder struct {
	rules     []string
	requests  []string
	responses []string
}

// The generic type parameter is a slice of bytes
func (b *BytesColumnBuilder) Build() Column {
	return Column{Rules: b.rules, Requests: b.requests, Responses: b.responses, ColType: "bytes"}
}

func BytesCol() *BytesColumnBuilder {
	return &BytesColumnBuilder{}
}

func (b *BytesColumnBuilder) Requests(r ...string) *BytesColumnBuilder {
	b.requests = append(b.requests, r...)
	return b
}

func (b *BytesColumnBuilder) Responses(r ...string) *BytesColumnBuilder {
	b.responses = append(b.responses, r...)
	return b
}

type TimeStampColumnBuilder struct {
	rules     []string
	requests  []string
	responses []string
}

func TimestampCol() *TimeStampColumnBuilder {
	return &TimeStampColumnBuilder{}
}

func (b *TimeStampColumnBuilder) Build() Column {
	return Column{Rules: b.rules, Requests: b.requests, Responses: b.responses, ColType: "timestamp"}
}

func (b *TimeStampColumnBuilder) Requests(r ...string) *TimeStampColumnBuilder {
	b.requests = append(b.requests, r...)
	return b
}

func (b *TimeStampColumnBuilder) Responses(r ...string) *TimeStampColumnBuilder {
	b.responses = append(b.responses, r...)
	return b
}
