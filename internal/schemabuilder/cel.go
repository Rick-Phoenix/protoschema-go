package schemabuilder

import (
	"errors"
	"fmt"
	"slices"
)

var DisableValidation = MessageOption{
	Name: "(buf.validate.message).disabled", Value: "true",
}

func GetCelOption(opt CelFieldOpts) string {
	return fmt.Sprintf("{\nid: %q \nmessage: %q\nexpression: %q\n}",
		opt.Id, opt.Message, opt.Expression)
}

func GetCelOptions(opts []CelFieldOpts) []string {
	flatOpts := []string{}

	for _, opt := range opts {
		stringOpt := fmt.Sprintf(
			"(buf.validate.field).cel = %s",
			GetCelOption(opt))

		flatOpts = append(flatOpts, stringOpt)
	}

	return flatOpts
}

type CelFieldOpts struct {
	Id, Message, Expression string
}

func GetOptions(optsMap map[string]any, repeatedOpts []string) ([]string, error) {
	flatOpts := []string{}
	var err error

	for name, value := range optsMap {
		val, fmtErr := GetProtoOption(name, value)

		if fmtErr != nil {
			err = errors.Join(err, fmtErr)
		}

		flatOpts = append(flatOpts, val)
	}

	flatOpts = slices.Concat(flatOpts, repeatedOpts)

	if err != nil {
		return flatOpts, err
	}

	return flatOpts, nil
}

func GetProtoOption(name string, value any) (string, error) {
	val, err := formatProtoValue(value)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s = %s", name, val), nil
}
