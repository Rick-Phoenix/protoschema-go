package schemabuilder

import (
	"errors"
	"fmt"
)

type OptionalField[BuilderT any] struct {
	internal *protoFieldInternal
	self     *BuilderT
}

func (of *OptionalField[BuilderT]) Optional() *BuilderT {
	if of.internal.required {
		of.internal.errors = errors.Join(of.internal.errors, fmt.Errorf("A field cannot be required and optional."))
	}
	of.internal.optional = true
	return of.self
}
