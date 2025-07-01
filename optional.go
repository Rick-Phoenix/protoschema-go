package schemabuilder

import u "github.com/Rick-Phoenix/goutils"

type OptionalField[BuilderT any] struct {
	optionalInternal *protoFieldInternal
	self             *BuilderT
}

func (of *OptionalField[BuilderT]) Optional() *BuilderT {
	of.optionalInternal.optional = true
	return of.self
}

func (of *OptionalField[BuilderT]) Nullable() *BuilderT {
	of.optionalInternal.goType = u.AddMissingPrefix(of.optionalInternal.goType, "*")
	return of.self
}
