package schemabuilder

import (
	"errors"
	"fmt"
	"strings"
)

type AnyField struct {
	*ProtoFieldExternal[AnyField, any]
	in    string
	notIn string
}

func ProtoAny(fieldNr uint) *AnyField {
	options := make(map[string]string)

	gf := &AnyField{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[AnyField, any]{&protoFieldInternal{fieldNr: fieldNr, protoType: "google.protobuf.Any", goType: "any", options: options, isNonScalar: true}, gf}

	return gf
}

func (af *AnyField) In(values ...string) *AnyField {
	list, err := formatProtoList(values)
	if err != nil {
		af.errors = errors.Join(af.errors, err)
	}
	af.in = list
	return af.self
}

func (af *AnyField) NotIn(values ...string) *AnyField {
	list, err := formatProtoList(values)
	if err != nil {
		af.errors = errors.Join(af.errors, err)
	}
	af.notIn = list
	return af.self
}

func (af *AnyField) Build(fieldName string, imports Set) (ProtoFieldData, error) {
	if af.errors != nil {
		return ProtoFieldData{}, af.errors
	}

	imports["google/protobuf/any.proto"] = present

	options := GetOptions(af.options, af.repeatedOptions)

	if af.in != "" || af.notIn != "" {
		var sb strings.Builder
		sb.WriteString("{\n")
		if af.in != "" {
			sb.WriteString(fmt.Sprintf("in: %s\n", af.in))
		}

		if af.notIn != "" {
			sb.WriteString(fmt.Sprintf("not_in: %s\n", af.notIn))
		}
		sb.WriteString("}")

		options = append(options, sb.String())
	}

	return ProtoFieldData{Name: fieldName, Options: options, ProtoType: af.protoType, GoType: af.goType, Optional: af.optional, FieldNr: af.fieldNr, Rules: af.rules, IsNonScalar: af.isNonScalar}, nil
}
