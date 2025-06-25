package schemabuilder

import (
	"errors"
	"fmt"
	"slices"
)

type ProtoMapBuilder struct {
	name     string
	keys     ProtoFieldBuilder
	values   ProtoFieldBuilder
	minPairs *uint
	maxPairs *uint
	*ProtoFieldExternal[ProtoMapBuilder]
}

func ProtoMap(name string, keys ProtoFieldBuilder, values ProtoFieldBuilder) *ProtoMapBuilder {
	options := make(map[string]any)
	rules := make(map[string]any)
	self := &ProtoMapBuilder{
		keys: keys, values: values, name: name,
	}

	self.ProtoFieldExternal = &ProtoFieldExternal[ProtoMapBuilder]{protoFieldInternal: &protoFieldInternal{
		options: options, rules: rules, isMap: true, isNonScalar: true, goType: fmt.Sprintf("map[%s]%s", keys.GetGoType(), values.GetGoType()),
	}, self: self}

	return self
}

func (b *ProtoMapBuilder) GetData() ProtoFieldData {
	data := b.protoFieldInternal.GetData()
	data.Name = b.name

	return data
}

func (b *ProtoMapBuilder) Build(fieldNr uint32, imports Set) (ProtoFieldData, error) {
	err := b.errors

	keysField, keysErr := b.keys.Build(fieldNr, imports)
	err = errors.Join(err, keysErr)

	valuesField, valsErr := b.values.Build(fieldNr, imports)
	err = errors.Join(err, valsErr)

	if !slices.Contains([]string{"string", "bool", "int32", "int64", "uint32", "uint64"}, keysField.ProtoType) {
		err = errors.Join(err, fmt.Errorf("Invalid type for a protobuf map key: '%s'", keysField.ProtoType))
	}

	if valuesField.Repeated {
		err = errors.Join(err, fmt.Errorf("Cannot use a repeated field as a value type in a protobuf map (must be wrapped in a message type first)."))
	}

	if valuesField.IsMap {
		err = errors.Join(err, fmt.Errorf("Cannot use a map as a value type of another map (must be wrapped in a message type first.)"))
	}

	options := []string{}

	for _, item := range []struct {
		MapType string
		Field   ProtoFieldData
	}{
		{"keys", keysField},
		{"values", valuesField},
	} {
		if len(item.Field.Rules) > 0 {
			rulesMap := make(map[string]any)
			rulesMap[item.Field.ProtoBaseType] = item.Field.Rules
			stringRule, fmtErr := formatProtoValue(rulesMap)
			err = errors.Join(err, fmtErr)

			options = append(options, fmt.Sprintf("(buf.validate.field).map.%s = %s", item.MapType, stringRule))
		}
	}

	extraOpts, optErr := GetOptions(b.options, b.repeatedOptions)

	options = append(options, extraOpts...)
	err = errors.Join(err, optErr)
	if err != nil {
		return ProtoFieldData{}, err
	}

	return ProtoFieldData{Name: b.name, ProtoType: fmt.Sprintf("map<%s, %s>", keysField.ProtoType, valuesField.ProtoType), GoType: b.goType, Optional: keysField.Optional, FieldNr: fieldNr, Options: options, IsNonScalar: true, IsMap: b.isMap}, nil
}

func (b *ProtoMapBuilder) MinPairs(n uint) *ProtoMapBuilder {
	if b.maxPairs != nil && *b.maxPairs < n {
		b.errors = errors.Join(b.errors, fmt.Errorf("min_pairs cannot be larger than max_pairs."))
	}
	b.options["(buf.validate.field).map.min_pairs"] = n
	b.minPairs = &n
	return b
}

func (b *ProtoMapBuilder) MaxPairs(n uint) *ProtoMapBuilder {
	if b.minPairs != nil && *b.minPairs > n {
		b.errors = errors.Join(b.errors, fmt.Errorf("min_pairs cannot be larger than max_pairs."))
	}
	b.options["(buf.validate.field).map.max_pairs"] = n
	b.maxPairs = &n
	return b
}
