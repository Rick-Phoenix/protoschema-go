package schemabuilder

import (
	"fmt"
	"strings"
)

type ProtoMapBuilder struct {
	keys     ProtoFieldBuilder
	values   ProtoFieldBuilder
	minPairs uint
	maxPairs uint
}

// Add cel and ignore options to this and others not implementing external
func ProtoMap(fieldNr uint, keys ProtoFieldBuilder, values ProtoFieldBuilder) *ProtoMapBuilder {
	return &ProtoMapBuilder{
		keys: keys, values: values,
	}
}

func (b *ProtoMapBuilder) Build(fieldName string, imports Set) (ProtoFieldData, error) {
	keysField, err := b.keys.Build(fieldName, imports)
	valuesField, err := b.values.Build(fieldName, imports)

	if keysField.Optional || valuesField.Optional {
		fmt.Printf("Ignoring 'optional' for map field %s...", fieldName)
	}

	options := []string{}

	if b.minPairs > 0 {
		options = append(options, fmt.Sprintf("(buf.validate.field).repeated.min_pairs = %d", b.minPairs))
	}

	if b.maxPairs > 0 {
		if b.maxPairs < b.minPairs {
			err = fmt.Errorf("- max_pairs cannot be smaller than min_pairs.\n%w", err)
		}

		options = append(options, fmt.Sprintf("(buf.validate.field).repeated.max_pairs = %d", b.minPairs))
	}

	if err != nil {
		return ProtoFieldData{}, err
	}

	for _, item := range []struct {
		Name  string
		Field ProtoFieldData
	}{
		{"keys", keysField},
		{"values", valuesField},
	} {
		if len(item.Field.Rules) > 0 {
			processedRules := 0
			stringRule := strings.Builder{}
			// Better formatting for this
			stringRule.WriteString(fmt.Sprintf("(buf.validate.field).map.%s = {\n", item.Name))
			stringRule.WriteString(fmt.Sprintf("  %s: {\n", item.Field.ProtoType))
			for name, value := range item.Field.Rules {
				if name == "required" {
					fmt.Printf("Ignoring 'required' for map key/value type %s...", fieldName)
					continue
				}

				stringValue, fmtErr := formatProtoValue(value)
				if fmtErr != nil {
					err = fmt.Errorf("- %s\n%w", fmtErr, err)
				} else {
					stringRule.WriteString(fmt.Sprintf("    %s: %s\n", name, stringValue))
					processedRules++
				}
			}
			stringRule.WriteString("}\n}")

			if processedRules > 0 {
				options = append(options, stringRule.String())
			}
		}
	}

	return ProtoFieldData{Name: fieldName, ProtoType: keysField.ProtoType, GoType: "[]" + keysField.GoType, Optional: keysField.Optional, FieldNr: keysField.FieldNr, Repeated: true, Options: options, IsNonScalar: true}, nil
}

func (b *ProtoMapBuilder) MinPairs(n uint) *ProtoMapBuilder {
	b.minPairs = n
	return b
}

func (b *ProtoMapBuilder) MaxPairs(n uint) *ProtoMapBuilder {
	b.maxPairs = n
	return b
}
