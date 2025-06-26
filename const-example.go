package schemabuilder

import (
	"errors"
	"fmt"
)

type FieldWithConst[BuilderT, ValueT any, SingleValT comparable] struct {
	constInternal *protoFieldInternal
	self          *BuilderT
	in            []SingleValT
	notIn         []SingleValT
}

func (b *FieldWithConst[BuilderT, ValueT, SingleValT]) Const(val ValueT) *BuilderT {
	b.constInternal.rules["const"] = val
	b.constInternal.isConst = true
	return b.self
}

func (b *FieldWithConst[BuilderT, ValueT, SingleValT]) Example(val ValueT) *BuilderT {
	opt, err := GetProtoOption("example", val)
	b.constInternal.errors = errors.Join(b.constInternal.errors, err)
	b.constInternal.repeatedOptions = append(b.constInternal.repeatedOptions, opt)
	return b.self
}

func (b *FieldWithConst[BuilderT, ValueT, SingleValT]) In(vals ...SingleValT) *BuilderT {
	if len(b.notIn) > 0 {
		overlaps := sliceIntersects(vals, b.notIn)
		if overlaps {
			b.constInternal.errors = errors.Join(b.constInternal.errors, fmt.Errorf("A field cannot be inside of 'in' and 'not_in' at the same time."))
		}

	}

	b.constInternal.rules["in"] = vals
	b.in = vals
	return b.self
}

func (b *FieldWithConst[BuilderT, ValueT, SingleValT]) NotIn(vals ...SingleValT) *BuilderT {
	if len(b.in) > 0 {
		overlaps := sliceIntersects(vals, b.in)

		if overlaps {
			b.constInternal.errors = errors.Join(b.constInternal.errors, fmt.Errorf("A field cannot be inside of 'in' and 'not_in' at the same time."))
		}
	}

	b.constInternal.rules["not_in"] = vals
	b.notIn = vals
	return b.self
}
