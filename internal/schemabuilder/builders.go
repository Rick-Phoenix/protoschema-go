package schemabuilder

import (
	"fmt"
	"time"
)

type TableSchema interface {
	Columns() Column[any]
}

type Column[T any] struct {
	Value     T
	Rules     []string
	Requests  []string
	Responses []string
}

// ColumnBuilder is an interface for any type that can produce a Column[T].
type ColumnBuilder[T any] interface {
	// Any type that has this method signature automatically implements the interface.
	Build() Column[T]
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

func (b *StringColumnBuilder) Build() Column[string] {
	return Column[string]{Rules: b.rules, Requests: b.requests, Responses: b.responses}
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
