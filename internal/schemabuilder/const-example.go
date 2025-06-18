package schemabuilder

import "fmt"

type FieldWithConst[BuilderT any, ValueT any] struct {
	internal *protoFieldInternal
	self     *BuilderT
}

func (b *FieldWithConst[BuilderT, ValueT]) Const(val ValueT) *BuilderT {
	formattedVal, err := formatProtoValue(val)
	if err != nil {
		b.internal.errors = append(b.internal.errors, err)
		return b.self
	}

	// For duration and timestam pthis should not be the entire name but only the last part
	b.internal.options[fmt.Sprintf("(buf.validate.field).%s.const", b.internal.protoType)] = formattedVal
	return b.self
}

func (b *FieldWithConst[BuilderT, ValueT]) Example(val ValueT) *BuilderT {
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
