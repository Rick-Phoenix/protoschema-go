package schemabuilder

import (
	"errors"
	"fmt"
	"strings"
)

type ProtoOneOfData struct {
	Name    string
	Choices []ProtoFieldData
	Options []ProtoOption
	Imports []string
}

type ProtoOneOfGroup struct {
	required bool
	name     string
	choices  ProtoOneOfsMap
	options  []ProtoOption
}

type ProtoOneOfBuilder interface {
	Build(imports Set) (ProtoOneOfData, error)
}

type ProtoOneOfsMap map[string]ProtoFieldBuilder

func ProtoOneOf(name string, choices ProtoOneOfsMap, options ...ProtoOption) *ProtoOneOfGroup {
	return &ProtoOneOfGroup{
		name: name, choices: choices, options: options,
	}
}

func (of *ProtoOneOfGroup) Build(imports Set) (ProtoOneOfData, error) {
	choicesData := []ProtoFieldData{}
	var fieldErr error

	for name, field := range of.choices {
		data, err := field.Build(name, imports)

		if err != nil {
			fieldErr = errors.Join(fieldErr, err)
		}

		if data.Optional {
			fmt.Printf("Ignoring 'optional' for member '%s' of a oneof group...\n", name)
		}

		if data.Repeated {
			fieldErr = errors.Join(fieldErr, fmt.Errorf("Cannot use a repeated field inside a oneof group."))
		}

		if strings.HasPrefix(data.ProtoType, "map<") {
			fieldErr = errors.Join(fieldErr, fmt.Errorf("Cannot use a map field inside a oneof group."))
		}

		choicesData = append(choicesData, data)
	}

	if fieldErr != nil {
		return ProtoOneOfData{}, fieldErr
	}

	return ProtoOneOfData{
		Name: of.name, Options: of.options, Choices: choicesData,
	}, nil
}

func (of *ProtoOneOfGroup) Required() *ProtoOneOfGroup {
	of.options = append(of.options, ProtoOption{
		Name:  "(buf.validate.oneof).required",
		Value: "true",
	})
	return of
}
