package schemabuilder

import "google.golang.org/protobuf/types/known/fieldmaskpb"

func FieldMask(name string) *GenericField {
	return MsgField(name, &MessageSchema{
		Name: "FieldMask", ImportPath: "google/protobuf/field_mask.proto", Model: &fieldmaskpb.FieldMask{},
		Package: &ProtoPackage{name: "google.protobuf"},
	})
}

func Empty() *MessageSchema {
	return &MessageSchema{Name: "Empty", ImportPath: "google/protobuf/empty.proto", Package: &ProtoPackage{name: "google.protobuf", goPackageName: "emptypb", goPackagePath: "google.golang.org/protobuf/types/known/emptypb"}}
}
