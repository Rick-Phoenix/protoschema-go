package schemabuilder

import "fmt"

// Column is the final data structure. It's simple and holds the results.
type Column[T any] struct {
	Value T // For our UnwrapToPlainStruct helper
	Rules []string
}

// ColumnBuilder is an interface for any type that can produce a Column[T].
type ColumnBuilder[T any] interface {
	// Any type that has this method signature automatically implements the interface.
	Build() Column[T]
}

type IntColumnBuilder struct {
	rules []string
}

func IntCol() *IntColumnBuilder { return &IntColumnBuilder{} }
func (b *IntColumnBuilder) GreaterThan(val int64) *IntColumnBuilder {
	b.rules = append(b.rules, fmt.Sprintf("(buf.validate.field).int64.gt = %d", val))
	return b
}
func (b *IntColumnBuilder) Build() Column[int] {
	return Column[int]{Rules: b.rules}
}

// StringColumnBuilder is a temporary object used to build a Column[string].
type StringColumnBuilder struct {
	rules []string
}

// StringCol is our "constructor" function. It's the entry point.
// It returns a pointer to the builder so we can chain methods.
func StringCol() *StringColumnBuilder {
	return &StringColumnBuilder{}
}

// --- Validation Methods ---

// Required adds the 'required' rule and returns the builder for chaining.
func (b *StringColumnBuilder) Required() *StringColumnBuilder {
	b.rules = append(b.rules, "(buf.validate.field).required = true")
	return b // Return self
}

// MinLen adds a minimum length rule.
func (b *StringColumnBuilder) MinLen(len uint64) *StringColumnBuilder {
	rule := fmt.Sprintf("(buf.validate.field).string.min_len = %d", len)
	b.rules = append(b.rules, rule)
	return b
}

// Email adds the 'email' format rule.
func (b *StringColumnBuilder) Email() *StringColumnBuilder {
	rule := fmt.Sprintf("(buf.validate.field).string.email = true")
	b.rules = append(b.rules, rule)
	return b
}

func (b *StringColumnBuilder) Build() Column[string] {
	return Column[string]{Rules: b.rules}
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
