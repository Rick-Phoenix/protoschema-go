package schemabuilder

import (
	"fmt"
	"strings"
)

type ProtoRepeatedBuilder struct {
	field    ProtoFieldBuilder
	unique   bool
	minItems uint
	maxItems uint
}

func RepeatedField(b ProtoFieldBuilder) *ProtoRepeatedBuilder {
	return &ProtoRepeatedBuilder{
		field: b,
	}
}

func (b *ProtoRepeatedBuilder) Build(fieldName string, imports Set) (ProtoFieldData, error) {
	fieldData, err := b.field.Build(fieldName, imports)

	if fieldData.Optional {
		err = fmt.Errorf("- A field cannot be optional and repeated.\n%w", err)
	}

	options := []string{}

	if b.unique {
		if fieldData.IsNonScalar {
			err = fmt.Errorf("- Cannot apply contraint 'unique' to a non-scalar repeated field.\n%w", err)
		}
		options = append(options, "(buf.validate.field).repeated.unique = true")
	}

	if b.minItems > 0 {
		options = append(options, fmt.Sprintf("(buf.validate.field).repeated.min_items = %d", b.minItems))
	}

	if b.maxItems > 0 {
		if b.maxItems < b.minItems {
			err = fmt.Errorf("- max_items cannot be smaller than min_items.\n%w", err)
		}

		options = append(options, fmt.Sprintf("(buf.validate.field).repeated.max_items = %d", b.minItems))
	}

	if err != nil {
		return ProtoFieldData{}, err
	}

	if len(fieldData.Rules) > 0 {
		processedRules := 0
		stringRule := strings.Builder{}
		// Better formatting for this
		stringRule.WriteString("(buf.validate.field).repeated.items = {\n")
		stringRule.WriteString(fmt.Sprintf("  %s: {\n", fieldData.ProtoType))
		for name, value := range fieldData.Rules {
			if name == "required" {
				fmt.Printf("Ignoring 'required' for repeated type %s...", fieldName)
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

	return ProtoFieldData{Name: fieldName, ProtoType: fieldData.ProtoType, GoType: "[]" + fieldData.GoType, Optional: fieldData.Optional, FieldNr: fieldData.FieldNr, Repeated: true, Options: options, IsNonScalar: true}, nil
}

func (b *ProtoRepeatedBuilder) Unique() *ProtoRepeatedBuilder {
	b.unique = true
	return b
}

func (b *ProtoRepeatedBuilder) MinItems(n uint) *ProtoRepeatedBuilder {
	b.minItems = n
	return b
}

func (b *ProtoRepeatedBuilder) MaxItems(n uint) *ProtoRepeatedBuilder {
	b.maxItems = n
	return b
}
