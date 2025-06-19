package schemabuilder

import (
	"errors"
	"fmt"
)

type FieldWithConst[BuilderT, ValueT any, SingleValT comparable] struct {
	internal *protoFieldInternal
	self     *BuilderT
	in       []SingleValT
	notIn    []SingleValT
}

func (b *FieldWithConst[BuilderT, ValueT, SingleValT]) Const(val ValueT) *BuilderT {
	formattedVal, err := formatProtoValue(val)
	if err != nil {
		b.internal.errors = errors.Join(b.internal.errors, err)
		return b.self
	}

	fieldName := b.internal.protoType
	switch b.internal.protoType {
	case "google.protobuf.Duration":
		fieldName = "duration"
	}

	b.internal.options[fmt.Sprintf("(buf.validate.field).%s.const", fieldName)] = formattedVal
	return b.self
}

func (b *FieldWithConst[BuilderT, ValueT, SingleValT]) Example(val ValueT) *BuilderT {
	formattedVal, err := formatProtoValue(val)
	if err != nil {
		b.internal.errors = errors.Join(b.internal.errors, err)
		return b.self
	}

	fieldName := b.internal.protoType
	switch b.internal.protoType {
	case "google.protobuf.Duration":
		fieldName = "duration"
	}

	b.internal.options[fmt.Sprintf("(buf.validate.field).%s.example", fieldName)] = formattedVal
	return b.self
}

func (b *FieldWithConst[BuilderT, ValueT, SingleValT]) In(vals ...SingleValT) *BuilderT {
	if len(b.notIn) > 0 {
		overlaps := SliceIntersects(vals, b.notIn)

		if overlaps {
			b.internal.errors = errors.Join(b.internal.errors, fmt.Errorf("A field cannot be inside of 'in' and 'not_in' at the same time."))
		}
	}
	list, err := formatProtoList(vals)
	if err != nil {
		b.internal.errors = errors.Join(b.internal.errors, err)
	}
	b.internal.rules["in"] = list
	return b.self
}

func (b *FieldWithConst[BuilderT, ValueT, SingleValT]) NotIn(vals ...SingleValT) *BuilderT {
	if len(b.in) > 0 {
		overlaps := SliceIntersects(vals, b.notIn)

		if overlaps {
			b.internal.errors = errors.Join(b.internal.errors, fmt.Errorf("A field cannot be inside of 'in' and 'not_in' at the same time."))
		}
	}
	list, err := formatProtoList(vals)
	if err != nil {
		b.internal.errors = errors.Join(b.internal.errors, err)
	}
	b.internal.rules["not_in"] = list
	return b.self
}
