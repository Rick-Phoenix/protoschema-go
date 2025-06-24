package schemabuilder

import (
	"errors"
	"maps"
	"slices"
)

type ProtoEnumMap map[int32]string

type ProtoEnumGroup struct {
	Name            string
	Members         ProtoEnumMap
	ReservedNames   []string
	ReservedNumbers []int32
	ReservedRanges  []Range
	Options         []ProtoOption
}

func ProtoEnum(name string, members ProtoEnumMap) ProtoEnumGroup {
	sorted := make(ProtoEnumMap)
	keys := slices.Sorted(maps.Keys(members))

	for _, k := range keys {
		v := members[k]
		sorted[k] = v
	}

	return ProtoEnumGroup{Name: name, Members: sorted}
}

func (e ProtoEnumGroup) Opts(o ...ProtoOption) ProtoEnumGroup {
	return ProtoEnumGroup{Name: e.Name, Members: e.Members, Options: o, ReservedNames: e.ReservedNames, ReservedNumbers: e.ReservedNumbers, ReservedRanges: e.ReservedRanges}
}

func (e ProtoEnumGroup) RsvNames(n ...string) ProtoEnumGroup {
	return ProtoEnumGroup{Name: e.Name, Members: e.Members, Options: e.Options, ReservedNames: n, ReservedNumbers: e.ReservedNumbers, ReservedRanges: e.ReservedRanges}
}

func (e ProtoEnumGroup) RsvNumbers(n ...int32) ProtoEnumGroup {
	return ProtoEnumGroup{Name: e.Name, Members: e.Members, Options: e.Options, ReservedNames: e.ReservedNames, ReservedNumbers: n, ReservedRanges: e.ReservedRanges}
}

func (e ProtoEnumGroup) RsvRanges(r ...Range) ProtoEnumGroup {
	return ProtoEnumGroup{Name: e.Name, Members: e.Members, Options: e.Options, ReservedNames: e.ReservedNames, ReservedNumbers: e.ReservedNumbers, ReservedRanges: r}
}

type EnumField struct {
	*ProtoFieldExternal[EnumField]
	*FieldWithConst[EnumField, int32, int32]
	*OptionalField[EnumField]
}

func ProtoEnumField(name string, enumName string) *EnumField {
	rules := make(map[string]any)
	options := make(map[string]any)

	ef := &EnumField{}
	internal := &protoFieldInternal{name: name, goType: "int32", protoType: enumName, rules: rules, isNonScalar: true, protoBaseType: "enum", options: options}

	ef.ProtoFieldExternal = &ProtoFieldExternal[EnumField]{
		protoFieldInternal: internal, self: ef,
	}
	ef.FieldWithConst = &FieldWithConst[EnumField, int32, int32]{constInternal: internal, self: ef}
	ef.OptionalField = &OptionalField[EnumField]{optionalInternal: internal, self: ef}

	return ef
}

func (ef *EnumField) Build(fieldNr uint32, imports Set) (ProtoFieldData, error) {
	data := ProtoFieldData{Name: ef.name, ProtoType: ef.protoType, GoType: ef.goType, FieldNr: fieldNr, Rules: ef.rules, IsNonScalar: false, Optional: ef.optional, ProtoBaseType: "enum"}

	var errAgg error
	errAgg = errors.Join(errAgg, ef.errors)

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
	errAgg = errors.Join(errAgg, err)

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

var AllowAlias = ProtoOption{Name: "allow_alias", Value: true}
