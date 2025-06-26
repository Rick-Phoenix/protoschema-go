package schemabuilder

type ProtoBoolField struct {
	*ProtoFieldExternal[ProtoBoolField]
	*FieldWithConst[ProtoBoolField, bool, bool]
	*OptionalField[ProtoBoolField]
}

func ProtoBool(name string) *ProtoBoolField {
	bf := &ProtoBoolField{}
	internal := &protoFieldInternal{
		name: name, protoType: "bool", goType: "bool",
	}
	bf.ProtoFieldExternal = &ProtoFieldExternal[ProtoBoolField]{internal, bf}
	bf.FieldWithConst = &FieldWithConst[ProtoBoolField, bool, bool]{constInternal: internal, self: bf}
	bf.OptionalField = &OptionalField[ProtoBoolField]{optionalInternal: internal, self: bf}

	return bf
}
