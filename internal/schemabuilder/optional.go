package schemabuilder

type OptionalField[BuilderT any] struct {
	internal *protoFieldInternal
	self     *BuilderT
}

func (of *OptionalField[BuilderT]) Optional() *BuilderT {
	of.internal.optional = true
	return of.self
}
