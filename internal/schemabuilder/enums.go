package schemabuilder

type ProtoEnumMap map[string]int32

type EnumField struct {
	*ProtoFieldExternal[EnumField, int32]
	*FieldWithConst[EnumField, int32, int32]
}

func ProtoEnumField(fieldNr uint, enumName string) *EnumField {
	ef := &EnumField{}
	internal := &protoFieldInternal{fieldNr: fieldNr, goType: "int32", protoType: "enumName"}
	ef.ProtoFieldExternal = &ProtoFieldExternal[EnumField, int32]{
		protoFieldInternal: internal, self: ef}
	ef.FieldWithConst = &FieldWithConst[EnumField, int32, int32]{constInternal: internal, self: ef}

	return ef
}

func (ef *EnumField) DefinedOnly() *EnumField {
	ef.rules["defined_only"] = true
	return ef
}
