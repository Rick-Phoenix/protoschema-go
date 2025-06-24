package schemabuilder

type AnyField struct {
	*ProtoFieldExternal[AnyField]
}

func ProtoAny(name string) *AnyField {
	options := make(map[string]any)
	rules := make(map[string]any)

	gf := &AnyField{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[AnyField]{&protoFieldInternal{name: name, protoType: "google.protobuf.Any", goType: "any", options: options, isNonScalar: true, rules: rules, imports: []string{"google/protobuf/any.proto"}}, gf}

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
