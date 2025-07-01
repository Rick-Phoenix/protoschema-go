package schemabuilder

type AnyField struct {
	*ProtoField[AnyField]
}

func Any(name string) *AnyField {
	options := make(map[string]any)
	rules := make(map[string]any)

	gf := &AnyField{}
	gf.ProtoField = &ProtoField[AnyField]{
		protoFieldInternal: &protoFieldInternal{
			name:        name,
			protoType:   "google.protobuf.Any",
			goType:      "any",
			options:     options,
			isNonScalar: true,
			rules:       rules,
			imports:     []string{"google/protobuf/any.proto"},
			messageRef: &MessageSchema{
				ImportPath: "google/protobuf/any.proto",
				Package: &ProtoPackage{
					GoPackagePath: "google.golang.org/protobuf/types/known/anypb",
					GoPackageName: "anypb",
					Name:          "Any",
				},
			},
		},
		self: gf,
	}

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
