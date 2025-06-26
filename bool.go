package schemabuilder

type ProtoBoolField struct {
	*ProtoFieldExternal[ProtoBoolField]
	*ProtoConstField[ProtoBoolField, bool, bool]
	*ProtoOptionalField[ProtoBoolField]
}

func Bool(name string) *ProtoBoolField {
	bf := &ProtoBoolField{}
	internal := &protoFieldInternal{
		name: name, protoType: "bool", goType: "bool",
	}
	bf.ProtoFieldExternal = &ProtoFieldExternal[ProtoBoolField]{internal, bf}
	bf.ProtoConstField = &ProtoConstField[ProtoBoolField, bool, bool]{constInternal: internal, self: bf}
	bf.ProtoOptionalField = &ProtoOptionalField[ProtoBoolField]{optionalInternal: internal, self: bf}

	return bf
}
