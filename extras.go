package schemabuilder

import "google.golang.org/protobuf/types/known/fieldmaskpb"

func FieldMask(name string) *GenericField {
	return MsgField(name, &MessageSchema{
		Name: "google.protobuf.FieldMask", ImportPath: "google/protobuf/field_mask.proto", Model: &fieldmaskpb.FieldMask{},
	})
}
