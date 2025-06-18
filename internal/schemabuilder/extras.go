package schemabuilder

type FieldMaskField struct {
	fieldNr uint
}

func FieldMask(fieldNr uint) *FieldMaskField {
	return &FieldMaskField{fieldNr: fieldNr}
}

func (fm *FieldMaskField) Build(fieldName string, imports Set) (ProtoFieldData, error) {

	imports["google/protobuf/field_mask.proto"] = present
	return ProtoFieldData{Name: fieldName, ProtoType: "google.protobuf.FieldMask", GoType: "*fieldmaskpb.FieldMask", Optional: false, FieldNr: fm.fieldNr, IsNonScalar: true}, nil
}
