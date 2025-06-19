package schemabuilder

import (
	"errors"
	"fmt"
	"maps"
	"strings"
)

type ProtoRepeatedBuilder struct {
	field    ProtoFieldBuilder
	unique   bool
	minItems uint
	maxItems uint
	fieldNr  uint
	*ProtoFieldExternal[ProtoRepeatedBuilder, any]
}

func RepeatedField(fieldNr uint, b ProtoFieldBuilder) *ProtoRepeatedBuilder {
	options := make(map[string]any)
	rules := make(map[string]any)
	self := &ProtoRepeatedBuilder{
		field: b, fieldNr: fieldNr,
	}

	self.ProtoFieldExternal = &ProtoFieldExternal[ProtoRepeatedBuilder, any]{protoFieldInternal: &protoFieldInternal{
		options: options, rules: rules,
	}, self: self}

	return self
}

func (b *ProtoRepeatedBuilder) Build(fieldName string, imports Set) (ProtoFieldData, error) {
	fieldData, err := b.field.Build(fieldName, imports)

	if fieldData.Optional {
		err = errors.Join(err, fmt.Errorf("A field cannot be optional and repeated."))
	}

	options := []string{}

	if b.unique {
		if fieldData.IsNonScalar {
			err = errors.Join(err, fmt.Errorf("Cannot apply contraint 'unique' to a non-scalar repeated field."))
		}
		options = append(options, "(buf.validate.field).repeated.unique = true")
	}

	if strings.HasPrefix(fieldData.ProtoType, "map<") {
		err = errors.Join(err, fmt.Errorf("Map fields cannot be repeated (must be wrapped in a message type)"))
	}

	if b.minItems > 0 {
		options = append(options, fmt.Sprintf("(buf.validate.field).repeated.min_items = %d", b.minItems))
	}

	if b.maxItems > 0 {
		if b.maxItems < b.minItems {
			err = errors.Join(err, fmt.Errorf("max_items cannot be smaller than min_items."))
		}

		options = append(options, fmt.Sprintf("(buf.validate.field).repeated.max_items = %d", b.minItems))
	}

	if fieldData.Repeated {
		err = errors.Join(err, fmt.Errorf("Cannot nest repeated fields inside one another (must be wrapped inside a message type first)"))
	}

	if fieldData.Required {
		fmt.Printf("Ignoring ineffective 'required' option for repeated field '%s' (you can set min_len to 1 instead to require at least one element)", fieldName)
	}

	if err != nil {
		return ProtoFieldData{}, err
	}

	if len(fieldData.Rules) > 0 {
		rulesMap := make(map[string]any)
		rulesCopy := make(map[string]any)
		maps.Copy(rulesCopy, fieldData.Rules)
		rulesMap[fieldData.ProtoType] = rulesCopy

		stringRules, fmtErr := formatProtoValue(rulesMap)
		if fmtErr != nil {
			err = errors.Join(err, fmtErr)
		}

		options = append(options, fmt.Sprintf("(buf.validate.field).repeated.items = %s", stringRules))
	}

	return ProtoFieldData{Name: fieldName, ProtoType: fieldData.ProtoType, GoType: "[]" + fieldData.GoType, Optional: fieldData.Optional, FieldNr: b.fieldNr, Repeated: true, Options: options, IsNonScalar: true}, nil
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
