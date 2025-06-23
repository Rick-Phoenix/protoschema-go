package schemabuilder

import "google.golang.org/protobuf/types/known/fieldmaskpb"

func FieldMask(name string) *GenericField[fieldmaskpb.FieldMask] {
	return MessageType[fieldmaskpb.FieldMask](name, "google.protobuf.FieldMask", "google/protobuf/field_mask.proto")
}
