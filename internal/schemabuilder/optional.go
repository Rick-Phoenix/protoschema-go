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
	if of.optionalInternal.required {
		of.optionalInternal.errors = errors.Join(of.optionalInternal.errors, fmt.Errorf("A field cannot be required and optional."))
	}
	of.optionalInternal.optional = true
	return of.self
}
