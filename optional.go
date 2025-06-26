package schemabuilder

type ProtoOptionalField[BuilderT any] struct {
	optionalInternal *protoFieldInternal
	self             *BuilderT
}

func (of *ProtoOptionalField[BuilderT]) Optional() *BuilderT {
	of.optionalInternal.optional = true
	return of.self
}
