package schemabuilder

import (
	"fmt"
)

type CelOption struct {
	Id, Message, Expression string
}

func NewCelOption(id, message, expression string) CelOption {
	return CelOption{Id: id, Message: message, Expression: expression}
}

func GetCelOption(opt CelOption) string {
	return fmt.Sprintf("{\nid: %q \nmessage: %q\nexpression: %q\n}",
		opt.Id, opt.Message, opt.Expression)
}

func GetCelOptions(opts []CelOption) []string {
	flatOpts := []string{}

	for _, opt := range opts {
		stringOpt := fmt.Sprintf(
			"(buf.validate.field).cel = %s",
			GetCelOption(opt))

		flatOpts = append(flatOpts, stringOpt)
	}

	return flatOpts
}
