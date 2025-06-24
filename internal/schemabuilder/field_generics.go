package schemabuilder

import (
	"errors"
)

func (b *ProtoFieldExternal[BuilderT, ValueT]) Options(o ...ProtoOption) *BuilderT {
	for _, op := range o {
		b.options[op.Name] = op.Value
	}
	return b.self
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) RepeatedOptions(o []ProtoOption) *BuilderT {
	var opts []string

	for _, v := range o {
		val, err := GetProtoOption(v.Name, v.Value)
		if err != nil {
			b.protoFieldInternal.errors = errors.Join(b.protoFieldInternal.errors, err)
		}

		opts = append(opts, val)
	}

	b.protoFieldInternal.repeatedOptions = append(b.protoFieldInternal.repeatedOptions, opts...)

	return b.self
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) IgnoreIfUnpopulated() *BuilderT {
	b.options["(buf.validate.field).ignore"] = "IGNORE_IF_UNPOPULATED"
	return b.self
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) IgnoreIfDefaultValue() *BuilderT {
	b.options["(buf.validate.field).ignore"] = "IGNORE_IF_DEFAULT_VALUE"
	return b.self
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) IgnoreAlways() *BuilderT {
	b.options["(buf.validate.field).ignore"] = "IGNORE_ALWAYS"
	return b.self
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) Deprecated() *BuilderT {
	b.options["deprecated"] = "true"
	return b.self
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) CelOptions(o ...CelFieldOpts) *BuilderT {
	opts := GetCelOptions(o)
	b.repeatedOptions = append(b.repeatedOptions, opts...)

	return b.self
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) Required() *BuilderT {
	b.options["(buf.validate.field).required"] = "true"
	b.required = true
	return b.self
}
