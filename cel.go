package schemabuilder

import "errors"

type CelOption struct {
	Id, Message, Expression string
}

func (b *ProtoFieldExternal[BuilderT]) CelOption(id, message, expression string) *BuilderT {
	opt, err := getProtoOption("(buf.validate.field).cel", CelOption{Id: id, Message: message, Expression: expression})
	b.errors = errors.Join(b.errors, err)
	b.repeatedOptions = append(b.repeatedOptions, opt)

	return b.self
}

func MessageCelOption(id, message, expression string) ProtoOption {
	out := ProtoOption{}

	out.Name = "(buf.validate.message).cel"

	out.Value = CelOption{Id: id, Message: message, Expression: expression}

	return out
}
