package schemabuilder

import "fmt"

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
	Build(fieldNr uint, imports Set) (ProtoOneOfData, Errors)
}

type ProtoOneOfsMap map[string]ProtoFieldBuilder

func ProtoOneOf(name string, choices ProtoOneOfsMap, options ...ProtoOption) *ProtoOneOfGroup {
	return &ProtoOneOfGroup{
		name: name, choices: choices, options: options,
	}
}

func (of *ProtoOneOfGroup) Build(name string, imports Set) (ProtoOneOfData, Errors) {
	choicesData := []ProtoFieldData{}
	var errors Errors

	for name, field := range of.choices {
		data, err := field.Build(name, imports)

		if err != nil {
			errors = append(errors, err...)
		}

		if data.Optional {
			errors = append(errors, fmt.Errorf("A field in a oneof group cannot be optional."))
		}

		choicesData = append(choicesData, data)
	}

	if errors != nil {
		return ProtoOneOfData{}, errors
	}

	return ProtoOneOfData{
		Name: name, Options: of.options, Choices: choicesData,
	}, []error{}
}

func (of *ProtoOneOfGroup) Required() *ProtoOneOfGroup {
	of.options = append(of.options, ProtoOption{
		Name:  "(buf.validate.oneof).required",
		Value: "true",
	})
	return of
}
