package schemabuilder

type ProtoBoolField struct {
	*ProtoFieldExternal[ProtoBoolField, bool]
	*FieldWithConst[ProtoBoolField, bool, bool]
	*OptionalField[ProtoBoolField]
}

func ProtoBool(fieldNr uint) *ProtoBoolField {
	bf := &ProtoBoolField{}
	internal := &protoFieldInternal{
		fieldNr: fieldNr, protoType: "bool", goType: "bool",
	}
	bf.ProtoFieldExternal = &ProtoFieldExternal[ProtoBoolField, bool]{internal, bf}
	bf.FieldWithConst = &FieldWithConst[ProtoBoolField, bool, bool]{constInternal: internal, self: bf}
	bf.OptionalField = &OptionalField[ProtoBoolField]{optionalInternal: internal, self: bf}

	return bf
}
