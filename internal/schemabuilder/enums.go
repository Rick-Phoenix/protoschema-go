package schemabuilder

type ProtoEnumMap map[string]int32

type EnumField struct {
	*ProtoFieldExternal[EnumField, int32]
	in    []int32
	notIn []int32
}

func ProtoEnumField(fieldNr uint, enumName string) *EnumField {
	ef := &EnumField{}
	ef.ProtoFieldExternal = &ProtoFieldExternal[EnumField, int32]{
		&protoFieldInternal{fieldNr: fieldNr, goType: "int32", protoType: "enumName"}, ef}

	return ef
}

func (ef *EnumField) DefinedOnly() *EnumField {
	ef.rules["defined_only"] = true
	return ef
}

func (ef *EnumField) In(vals ...int32) *EnumField {
	list, err := formatProtoList(vals)
	if err != nil {
		ef.errors = append(ef.errors, err)
	}
	ef.rules["in"] = list
	return ef.self
}

func (ef *EnumField) NotIn(vals ...int32) *EnumField {
	list, err := formatProtoList(vals)
	if err != nil {
		ef.errors = append(ef.errors, err)
	}
	ef.rules["not_in"] = list
	return ef.self
}
