package schemabuilder

import (
	"errors"
	"maps"
)

type ProtoEnumMap map[string]int32

type ProtoEnumGroup struct {
	Name            string
	Members         ProtoEnumMap
	ReservedNames   []string
	ReservedNumbers []int32
	Options         []ProtoOption
}

func ProtoEnum(name string, members ProtoEnumMap) ProtoEnumGroup {
	return ProtoEnumGroup{Name: name, Members: members}
}

func (e ProtoEnumGroup) SetOptions(o ...ProtoOption) ProtoEnumGroup {
	return ProtoEnumGroup{Name: e.Name, Members: e.Members, Options: o, ReservedNames: e.ReservedNames, ReservedNumbers: e.ReservedNumbers}
}

func (e ProtoEnumGroup) SetReservedNames(n ...string) ProtoEnumGroup {
	return ProtoEnumGroup{Name: e.Name, Members: e.Members, Options: e.Options, ReservedNames: n, ReservedNumbers: e.ReservedNumbers}
}

func (e ProtoEnumGroup) SetReservedNumbers(n ...int32) ProtoEnumGroup {
	return ProtoEnumGroup{Name: e.Name, Members: e.Members, Options: e.Options, ReservedNames: e.ReservedNames, ReservedNumbers: n}
}

type EnumField struct {
	*ProtoFieldExternal[EnumField, int32]
	*FieldWithConst[EnumField, int32, int32]
	*OptionalField[EnumField]
}

func ProtoEnumField(fieldNr uint, enumName string) *EnumField {
	rules := make(map[string]any)
	ef := &EnumField{}
	internal := &protoFieldInternal{fieldNr: fieldNr, goType: "int32", protoType: enumName, rules: rules, isNonScalar: true, protoBaseType: "enum"}
	ef.ProtoFieldExternal = &ProtoFieldExternal[EnumField, int32]{
		protoFieldInternal: internal, self: ef}
	ef.FieldWithConst = &FieldWithConst[EnumField, int32, int32]{constInternal: internal, self: ef}
	ef.OptionalField = &OptionalField[EnumField]{optionalInternal: internal, self: ef}

	return ef
}

func (ef *EnumField) Build(fieldName string, imports Set) (ProtoFieldData, error) {
	data := ProtoFieldData{Name: fieldName, ProtoType: ef.protoType, GoType: ef.goType, FieldNr: ef.fieldNr, Rules: ef.rules, IsNonScalar: false, Optional: ef.optional, ProtoBaseType: "enum"}

	var errAgg error
	if ef.errors != nil {
		errAgg = errors.Join(errAgg, ef.errors)
	}

	for _, v := range ef.imports {
		imports[v] = present
	}

	var options []string

	optsCollector := make(map[string]any)
	maps.Copy(optsCollector, ef.options)

	if len(ef.rules) > 0 {
		imports["buf/validate/validate.proto"] = present
		enumRules := make(map[string]any)

		if list, exists := ef.rules["in"]; exists {
			enumRules["in"] = list
		}

		if list, exists := ef.rules["not_in"]; exists {
			enumRules["not_in"] = list
		}

		optsCollector["(buf.validate.field).enum"] = enumRules
	}

	options, err := GetOptions(optsCollector, ef.repeatedOptions)
	if err != nil {
		errAgg = errors.Join(errAgg, err)
	}

	data.Options = options

	if errAgg != nil {
		return data, errAgg
	}

	return data, nil
}

func (ef *EnumField) DefinedOnly() *EnumField {
	ef.rules["defined_only"] = true
	return ef
}
