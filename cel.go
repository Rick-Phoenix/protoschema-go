package protoschema

import "errors"

// A custom validator setting for protovalidate. See the protovalidate docs to learn more about the usage for these.
type CelOption struct {
	// The identifier for this rule
	Id,
	// The message to display when validation fails
	Message,
	// The CEL expression that validates the field
	Expression string
}

// A method to add a Cel option to a specific field.
func (b *ProtoField[BuilderT]) CelOption(id, message, expression string) *BuilderT {
	opt, err := getProtoOption("(buf.validate.field).cel", CelOption{Id: id, Message: message, Expression: expression})
	b.errors = errors.Join(b.errors, err)
	b.repeatedOptions = append(b.repeatedOptions, opt)

	return b.self
}
