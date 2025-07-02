package protoschema

// An instance of a google.protobuf.Any field.
type AnyField struct {
	*ProtoField[AnyField]
}

// The constructor for a google.protobuf.Any field.
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

// Rule: this field's type_url must be among those listed in order to be accepted.
func (af *AnyField) In(typeUrls ...string) *AnyField {
	af.protoFieldInternal.rules["in"] = typeUrls
	return af.self
}

// Rule: this Any field's type_url must not be among those listed in order to be accepted.
func (af *AnyField) NotIn(typeUrls ...string) *AnyField {
	af.protoFieldInternal.rules["not_in"] = typeUrls
	return af.self
}
