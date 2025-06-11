package schemabuilder

import (
	"fmt"
	"time"
)

type ColumnsMap map[string]ColumnBuilder2

type TableSchema interface {
	GetColumns() ColumnsMap
}

type TableBuilder struct {
	Name    string
	Columns ColumnsMap
}

func (b *TableBuilder) GetColumns() ColumnsMap {
	return b.Columns
}

var User2 = &TableBuilder{
	Name:    "User",
	Columns: ColumnsMap{"Name": StringCol().Required().MinLen(3).Requests("create").Responses("get", "create")},
}

func DoThings(t TableSchema) bool {
	return true
}

var test = DoThings(User2)

type Column2 struct {
	Rules     []string
	Requests  []string
	Responses []string
	ColType   string
}

type Column[T any] struct {
	Value     T
	Rules     []string
	Requests  []string
	Responses []string
}

type ColumnBuilder2 interface {
	Build() Column2
}

// ColumnBuilder is an interface for any type that can produce a Column[T].
type ColumnBuilder[T any] interface {
	// Any type that has this method signature automatically implements the interface.
	Build() Column[T]
}

type UserSchema struct {
	// ID          ColumnBuilder[int64]  `bun:"id,pk,autoincrement" json:"id"`
	Name  ColumnBuilder[string] `bun:"name,notnull" json:"user_name"`
	Email ColumnBuilder[string] `bun:"email,unique"`
	// Age         ColumnBuilder[int64]
}

// var UserExample = UserSchema{
// 	// This works because the value returned by StringCol().Required().MinLen(3)
// 	// is a *StringColumnBuilder, which satisfies the ColumnBuilder[string] interface.
// 	Name: StringCol().Required().MinLen(3).Requests("create").Responses("get", "create"),
//
// 	Email: StringCol().Required().Email().Requests("create").Responses("get"),
//
// 	// Age: Int64Col(),
// }

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

func (b *StringColumnBuilder) Build() Column2 {
	return Column2{Rules: b.rules, Requests: b.requests, Responses: b.responses, ColType: "string"}
}

type Int64ColumnBuilder struct {
	rules []string
}

func Int64Col() *Int64ColumnBuilder {
	return &Int64ColumnBuilder{}
}

func (b *Int64ColumnBuilder) Build() Column[int64] {
	return Column[int64]{Rules: b.rules}
}

type BytesColumnBuilder struct {
	rules []string
}

// The generic type parameter is a slice of bytes
func (b *BytesColumnBuilder) Build() Column[[]byte] {
	return Column[[]byte]{Rules: b.rules}
}

func BytesCol() *BytesColumnBuilder {
	return &BytesColumnBuilder{}
}

type TimeStampColumnBuilder struct {
	rules []string
}

func (b *TimeStampColumnBuilder) Build() Column[time.Time] {
	return Column[time.Time]{Rules: b.rules}
}
