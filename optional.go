package schemabuilder

import u "github.com/Rick-Phoenix/goutils"

type OptionalField[BuilderT any] struct {
	optionalInternal *protoFieldInternal
	self             *BuilderT
}

func (of *OptionalField[BuilderT]) clone(internalClone *protoFieldInternal, selfClone *BuilderT) *OptionalField[BuilderT] {
	clone := *of
	clone.optionalInternal = internalClone
	clone.self = selfClone
	return &clone
}

func (of *OptionalField[BuilderT]) Optional() *BuilderT {
	of.optionalInternal.optional = true
	return of.self
}

func (of *OptionalField[BuilderT]) Nullable() *BuilderT {
	of.optionalInternal.goType = u.AddMissingPrefix(of.optionalInternal.goType, "*")
	return of.self
}
