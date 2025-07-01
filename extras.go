package schemabuilder

import "google.golang.org/protobuf/types/known/fieldmaskpb"

// A constructor for a google.protobuf.FieldMask protobuf field.
func FieldMask(name string) *GenericField {
	return MsgField(name, &MessageSchema{
		Name:       "FieldMask",
		ImportPath: "google/protobuf/field_mask.proto",
		Model:      &fieldmaskpb.FieldMask{},
		Package: &ProtoPackage{
			Name:          "google.protobuf",
			GoPackagePath: "google.golang.org/protobuf/types/known/fieldmaskpb",
		},
	})
}

// A constructor for a google.protobuf.Empty protobuf field.
func Empty() *MessageSchema {
	return &MessageSchema{
		Name:       "Empty",
		ImportPath: "google/protobuf/empty.proto",
		Package: &ProtoPackage{
			Name:          "google.protobuf",
			GoPackagePath: "google.golang.org/protobuf/types/known/emptypb",
		},
	}
}
