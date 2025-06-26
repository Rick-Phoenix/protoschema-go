package schemabuilder

import (
	"errors"
	"maps"
)

type EnumMembers map[int32]string

type EnumGroup struct {
	Name            string
	Members         EnumMembers
	ReservedNames   []string
	ReservedNumbers []int32
	ReservedRanges  []Range
	Options         []ProtoOption
}

type EnumField struct {
	*ProtoField[EnumField]
	*ConstField[EnumField, int32, int32]
	*OptionalField[EnumField]
}

func Enum(name string, enumName string) *EnumField {
	rules := make(map[string]any)
	options := make(map[string]any)

	ef := &EnumField{}
	internal := &protoFieldInternal{name: name, goType: "int32", protoType: enumName, rules: rules, protoBaseType: "enum", options: options}

	ef.ProtoField = &ProtoField[EnumField]{
		protoFieldInternal: internal, self: ef,
	}
	ef.ConstField = &ConstField[EnumField, int32, int32]{constInternal: internal, self: ef}
	ef.OptionalField = &OptionalField[EnumField]{optionalInternal: internal, self: ef}

	return ef
}

func (ef *EnumField) Build(fieldNr uint32, imports Set) (FieldData, error) {
	data := FieldData{Name: ef.name, ProtoType: ef.protoType, GoType: ef.goType, FieldNr: fieldNr, Rules: ef.rules, Optional: ef.optional, ProtoBaseType: "enum"}

	var errAgg error
	errAgg = errors.Join(errAgg, ef.errors)

	options := make([]string, len(ef.repeatedOptions))
	copy(options, ef.repeatedOptions)

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

	options, err := getOptions(optsCollector, options)
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
