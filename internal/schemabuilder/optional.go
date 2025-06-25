package schemabuilder

import (
	"errors"
	"fmt"
)

type OptionalField[BuilderT any] struct {
	optionalInternal *protoFieldInternal
	self             *BuilderT
}

func (of *OptionalField[BuilderT]) Optional() *BuilderT {
	if of.optionalInternal.isConst {
		of.optionalInternal.errors = errors.Join(of.optionalInternal.errors, fmt.Errorf("A constant field cannot be optional."))
	}
	of.optionalInternal.optional = true
	return of.self
}
