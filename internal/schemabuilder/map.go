package schemabuilder

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

type ProtoMapBuilder struct {
	keys     ProtoFieldBuilder
	values   ProtoFieldBuilder
	minPairs *uint
	maxPairs *uint
	fieldNr  uint
	*ProtoFieldExternal[ProtoMapBuilder, any]
}

// Add cel and ignore options to this and others not implementing external
func ProtoMap(fieldNr uint, keys ProtoFieldBuilder, values ProtoFieldBuilder) *ProtoMapBuilder {
	options := make(map[string]any)
	rules := make(map[string]any)
	self := &ProtoMapBuilder{
		keys: keys, values: values, fieldNr: fieldNr,
	}

	self.ProtoFieldExternal = &ProtoFieldExternal[ProtoMapBuilder, any]{protoFieldInternal: &protoFieldInternal{
		options: options, rules: rules,
	}, self: self}

	return self
}

func (b *ProtoMapBuilder) Build(fieldName string, imports Set) (ProtoFieldData, error) {
	var err error

	keysField, keysErr := b.keys.Build(fieldName, imports)

	if keysErr != nil {
		err = errors.Join(err, keysErr)
	}

	valuesField, valsErr := b.values.Build(fieldName, imports)

	if valsErr != nil {
		err = errors.Join(err, valsErr)
	}

	if !slices.Contains([]string{"string", "bool", "int32", "int64", "uint32", "uint64"}, keysField.ProtoType) {
		err = errors.Join(err, fmt.Errorf("Invalid type for a protobuf map key: '%s'", keysField.ProtoType))
	}

	if valuesField.Repeated {
		err = errors.Join(err, fmt.Errorf("Cannot use a repeated field as a value type in a protobuf map (must be wrapped in a message type first)."))
	}

	if strings.HasPrefix(valuesField.ProtoType, "map<") {
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
			stringRule, fmtErr := formatProtoValue(item.Field.Rules)
			if fmtErr != nil {
				err = errors.Join(err, fmtErr)
			}

			options = append(options, fmt.Sprintf("(buf.validate.field).map.%s = %s", item.MapType, stringRule))
		}
	}

	extraOpts, optErr := GetOptions(b.options, b.repeatedOptions)

	options = append(options, extraOpts...)

	if optErr != nil {
		err = errors.Join(err, optErr)
	}

	if err != nil {
		return ProtoFieldData{}, err
	}

	return ProtoFieldData{Name: fieldName, ProtoType: fmt.Sprintf("map<%s, %s>", keysField.ProtoType, valuesField.ProtoType), GoType: "[]" + keysField.GoType, Optional: keysField.Optional, FieldNr: b.fieldNr, Options: options, IsNonScalar: true}, nil
}

func (b *ProtoMapBuilder) MinPairs(n uint) *ProtoMapBuilder {
	if b.maxPairs != nil && *b.maxPairs < n {
		b.errors = errors.Join(b.errors, fmt.Errorf("min_pairs cannot be larger than max_pairs."))
	}
	b.options["(buf.validate.field).repeated.min_pairs"] = n
	b.minPairs = &n
	return b
}

func (b *ProtoMapBuilder) MaxPairs(n uint) *ProtoMapBuilder {
	if b.minPairs != nil && *b.minPairs > n {
		b.errors = errors.Join(b.errors, fmt.Errorf("min_pairs cannot be larger than max_pairs."))
	}
	b.options["(buf.validate.field).repeated.max_pairs"] = n
	b.maxPairs = &n
	return b
}
