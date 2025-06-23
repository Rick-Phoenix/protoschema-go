package schemabuilder

import (
	"errors"
)

type AnyField struct {
	*ProtoFieldExternal[AnyField, any]
}

func ProtoAny(name string) *AnyField {
	options := make(map[string]any)
	rules := make(map[string]any)

	gf := &AnyField{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[AnyField, any]{&protoFieldInternal{name: name, protoType: "google.protobuf.Any", goType: "any", options: options, isNonScalar: true, rules: rules}, gf}

	return gf
}

func (af *AnyField) In(values ...string) *AnyField {
	af.protoFieldInternal.rules["in"] = values
	return af.self
}

func (af *AnyField) NotIn(values ...string) *AnyField {
	af.protoFieldInternal.rules["not_in"] = values
	return af.self
}

func (af *AnyField) Build(fieldName string, imports Set) (ProtoFieldData, error) {
	var errAgg error

	if af.errors != nil {
		errAgg = errors.Join(errAgg, af.errors)
	}

	imports["google/protobuf/any.proto"] = present

	options, err := GetOptions(af.options, af.repeatedOptions)

	if err != nil {
		errAgg = errors.Join(errAgg, err)
	}

	return ProtoFieldData{Name: fieldName, Options: options, ProtoType: af.protoType, GoType: af.goType, Optional: af.optional, FieldNr: af.fieldNr, Rules: af.rules, IsNonScalar: af.isNonScalar}, nil
}
