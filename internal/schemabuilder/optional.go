package schemabuilder

type OptionalField[BuilderT any] struct {
	optionalInternal *protoFieldInternal
	self             *BuilderT
}

func (of *OptionalField[BuilderT]) Optional() *BuilderT {
	of.optionalInternal.optional = true
	return of.self
}
