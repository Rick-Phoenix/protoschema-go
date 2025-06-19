package schemabuilder

import (
	"errors"
	"fmt"
)

func (b *ProtoFieldExternal[BuilderT, ValueT]) Options(o []ProtoOption) *BuilderT {
	for _, op := range o {
		b.options[op.Name] = op.Value
	}
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

func (b *ProtoFieldExternal[BuilderT, ValueT]) CelOption(o CelFieldOpts) *BuilderT {
	b.repeatedOptions = append(b.repeatedOptions, GetCelOption(CelFieldOpts{
		Id: o.Id, Expression: o.Expression, Message: o.Message,
	}))

	return b.self
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) Required() *BuilderT {
	if b.optional {
		b.errors = errors.Join(b.errors, fmt.Errorf("A field cannot be required and optional."))
	}
	b.options["(buf.validate.field).required"] = "true"
	b.required = true
	return b.self
}
