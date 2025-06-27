package schemabuilder

func FieldMask(name string) *GenericField {
	return MsgField(name, &MessageSchema{
		Name: "google.protobuf.FieldMask", ImportPath: "google/protobuf/field_mask.proto",
	})
}
