package protoschema

import u "github.com/Rick-Phoenix/goutils"

// A subtype for a protobuf field that allows the use of the optional keyword.
type OptionalField[BuilderT any] struct {
	optionalInternal *protoFieldInternal
	self             *BuilderT
}

// Sets a field as optional.
func (of *OptionalField[BuilderT]) Optional() *BuilderT {
	of.optionalInternal.optional = true
	return of.self
}

// Sets a field as nullable. This does not affect the output of the proto file and only affects model validation.
// Calling this method means that the model validator will expect the corresponding field in the model to be a pointer.
func (of *OptionalField[BuilderT]) Nullable() *BuilderT {
	of.optionalInternal.goType = u.AddMissingPrefix(of.optionalInternal.goType, "*")
	return of.self
}
