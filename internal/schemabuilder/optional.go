package schemabuilder

import "strings"

type OptionalField[BuilderT any] struct {
	optionalInternal *protoFieldInternal
	self             *BuilderT
}

func (of *OptionalField[BuilderT]) Optional() *BuilderT {
	if !strings.HasPrefix("*", of.optionalInternal.goType) {
		of.optionalInternal.goType = "*" + of.optionalInternal.goType
	}
	of.optionalInternal.optional = true
	return of.self
}
