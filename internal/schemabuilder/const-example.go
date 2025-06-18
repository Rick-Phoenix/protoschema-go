package schemabuilder

import (
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
		b.internal.errors = append(b.internal.errors, err)
		return b.self
	}

	// For duration and timestam pthis should not be the entire name but only the last part
	b.internal.options[fmt.Sprintf("(buf.validate.field).%s.const", b.internal.protoType)] = formattedVal
	return b.self
}

func (b *FieldWithConst[BuilderT, ValueT, SingleValT]) Example(val ValueT) *BuilderT {
	if b.internal.protoType == "any" {
		b.internal.errors = append(b.internal.errors, fmt.Errorf("Method 'Example()' is not supposed for google.protobuf.Any."))
	}
	formattedVal, err := formatProtoValue(val)
	if err != nil {
		b.internal.errors = append(b.internal.errors, err)
		return b.self
	}

	// Make this repeatable
	b.internal.repeatedOptions = append(b.internal.repeatedOptions, fmt.Sprintf("(buf.validate.field).%s.example = %s", b.internal.protoType, formattedVal))
	return b.self
}

func (b *FieldWithConst[BuilderT, ValueT, SingleValT]) In(vals ...SingleValT) *BuilderT {
	if len(b.notIn) > 0 {
		overlaps := SliceIntersects(vals, b.notIn)

		if overlaps {
			b.internal.errors = append(b.internal.errors, fmt.Errorf("A field cannot be inside of 'in' and 'not_in' at the same time."))
		}
	}
	list, err := formatProtoList(vals)
	if err != nil {
		b.internal.errors = append(b.internal.errors, err)
	}
	b.internal.rules["in"] = list
	return b.self
}

func (b *FieldWithConst[BuilderT, ValueT, SingleValT]) NotIn(vals ...SingleValT) *BuilderT {
	if len(b.in) > 0 {
		overlaps := SliceIntersects(vals, b.notIn)

		if overlaps {
			b.internal.errors = append(b.internal.errors, fmt.Errorf("A field cannot be inside of 'in' and 'not_in' at the same time."))
		}
	}
	list, err := formatProtoList(vals)
	if err != nil {
		b.internal.errors = append(b.internal.errors, err)
	}
	b.internal.rules["not_in"] = list
	return b.self
}
