package schemabuilder

import "google.golang.org/protobuf/types/known/fieldmaskpb"

func FieldMask(fieldNr uint) *GenericField[fieldmaskpb.FieldMask] {
	return MessageType[fieldmaskpb.FieldMask](fieldNr, "google.protobuf.FieldMask", WithImportPath("google/protobuf/field_mask.proto"))
}
