package schemabuilder

import (
	"fmt"
	"maps"
	"slices"
)

func MessageCelOption(o CelFieldOpts) MessageOption {
	return MessageOption{
		Name: "(buf.validate.field).cel", Value: GetCelOption(o),
	}
}

var DisableValidation = MessageOption{
	Name: "(buf.validate.message).disabled", Value: "true",
}

func GetCelOption(opt CelFieldOpts) string {
	return fmt.Sprintf(
		`{
		id: %q
		message: %q
		expression: %q
		}`,
		opt.Id, opt.Message, opt.Expression)

}

func GetCelOptions(opts []CelFieldOpts) []string {
	flatOpts := []string{}

	for i, opt := range opts {
		stringOpt := fmt.Sprintf(
			`(buf.validate.field).cel = %s`,
			GetCelOption(opt))
		if i < len(opts)-1 {
			stringOpt += ", "
		}

		flatOpts = append(flatOpts, stringOpt)
	}

	return flatOpts
}

type CelFieldOpts struct {
	Id, Message, Expression string
}

func GetOptions(optsMap map[string]string, repeatedOpts []string) []string {
	flatOpts := []string{}
	optNames := slices.Collect(maps.Keys(optsMap))

	for _, name := range optNames {
		stringOpt := name + " = " + optsMap[name]

		flatOpts = append(flatOpts, stringOpt)
	}

	flatOpts = slices.Concat(flatOpts, repeatedOpts)

	return flatOpts
}
