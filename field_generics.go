package protoschema

import (
	"errors"
)

// The options to add to this protobuf field. Use RepeatedOptions for repeated options.
func (b *ProtoField[BuilderT]) Options(o ...ProtoOption) *BuilderT {
	for _, op := range o {
		b.options[op.Name] = op.Value
	}
	return b.self
}

// The repeated options to add to this protobuf field.
func (b *ProtoField[BuilderT]) RepeatedOptions(o ...ProtoOption) *BuilderT {
	var opts []string

	for _, v := range o {
		val, err := getProtoOption(v.Name, v.Value)
		if err != nil {
			b.protoFieldInternal.errors = errors.Join(b.protoFieldInternal.errors, err)
		}

		opts = append(opts, val)
	}

	b.protoFieldInternal.repeatedOptions = append(b.protoFieldInternal.repeatedOptions, opts...)

	return b.self
}

// Skips validation if the value is nullable and unpopulated.
func (b *ProtoField[BuilderT]) IgnoreIfUnspecified() *BuilderT {
	b.options["(buf.validate.field).ignore"] = "IGNORE_UNSPECIFIED"
	return b.self
}

// Skips validation if the field is unset.
func (b *ProtoField[BuilderT]) IgnoreIfUnpopulated() *BuilderT {
	b.options["(buf.validate.field).ignore"] = "IGNORE_IF_UNPOPULATED"
	return b.self
}

// Skips validation if the field's value is its default value.
func (b *ProtoField[BuilderT]) IgnoreIfDefaultValue() *BuilderT {
	b.options["(buf.validate.field).ignore"] = "IGNORE_IF_DEFAULT_VALUE"
	return b.self
}

// Turns off validation for a field.
func (b *ProtoField[BuilderT]) IgnoreAlways() *BuilderT {
	b.options["(buf.validate.field).ignore"] = "IGNORE_ALWAYS"
	return b.self
}

// Marks the field as deprecated.
func (b *ProtoField[BuilderT]) Deprecated() *BuilderT {
	b.options["deprecated"] = true
	return b.self
}

// Rule: this field is required. This means that:
// 1. If the field has a message type, is optional, or is part of a oneof group, it must be explicitely set (even to its default value)
// 2. If the field is non-scalar, it cannot be its default value.
// 3. If the field is repeated, or a map, it must have a length of at least one.
func (b *ProtoField[BuilderT]) Required() *BuilderT {
	b.options["(buf.validate.field).required"] = true
	b.required = true
	return b.self
}
