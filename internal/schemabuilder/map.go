package schemabuilder

import (
	"errors"
	"fmt"
	"slices"
)

type ProtoMapBuilder struct {
	keys     ProtoFieldBuilder
	values   ProtoFieldBuilder
	minPairs uint
	maxPairs uint
	fieldNr  uint
}

// Add cel and ignore options to this and others not implementing external
func ProtoMap(fieldNr uint, keys ProtoFieldBuilder, values ProtoFieldBuilder) *ProtoMapBuilder {
	return &ProtoMapBuilder{
		keys: keys, values: values, fieldNr: fieldNr,
	}
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
		err = errors.Join(err, fmt.Errorf("Cannot use a repeated field as a value type in a protobuf map (must be wrapped in a message type)."))
	}

	if keysField.Optional || valuesField.Optional {
		fmt.Printf("Ignoring 'optional' for map field '%s' (use min_pairs to require at least one element)", fieldName)
	}

	options := []string{}

	if b.minPairs > 0 {
		options = append(options, fmt.Sprintf("(buf.validate.field).repeated.min_pairs = %d", b.minPairs))
	}

	if b.maxPairs > 0 {
		if b.maxPairs < b.minPairs {
			err = errors.Join(err, fmt.Errorf("max_pairs cannot be smaller than min_pairs."))
		}

		options = append(options, fmt.Sprintf("(buf.validate.field).repeated.max_pairs = %d", b.minPairs))
	}

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

	if err != nil {
		return ProtoFieldData{}, err
	}

	return ProtoFieldData{Name: fieldName, ProtoType: fmt.Sprintf("map<%s, %s>", keysField.ProtoType, valuesField.ProtoType), GoType: "[]" + keysField.GoType, Optional: keysField.Optional, FieldNr: b.fieldNr, Options: options, IsNonScalar: true}, nil
}

func (b *ProtoMapBuilder) MinPairs(n uint) *ProtoMapBuilder {
	b.minPairs = n
	return b
}

func (b *ProtoMapBuilder) MaxPairs(n uint) *ProtoMapBuilder {
	b.maxPairs = n
	return b
}
