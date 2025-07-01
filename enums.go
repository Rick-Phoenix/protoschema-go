package schemabuilder

import (
	"errors"
	"maps"

	"github.com/labstack/gommon/log"
)

type EnumMembers map[int32]string

type EnumGroup struct {
	Name            string
	Members         EnumMembers
	ReservedNames   []string
	ReservedNumbers []int32
	ReservedRanges  []Range
	Options         []ProtoOption
	Package         *ProtoPackage
	File            *FileSchema
	Message         *MessageSchema
	Metadata        map[string]any
	ImportPath      string
}

func (e *EnumGroup) GetName() string {
	if e == nil {
		return ""
	}

	name := e.Name
	if e.Message != nil {
		name = e.Message.GetName() + "." + e.Name
	}

	return name
}

func (e *EnumGroup) GetFullName(p *ProtoPackage) string {
	if e == nil {
		return ""
	}

	if e.Package == nil || e.Package == p {
		return e.GetName()
	}

	return e.Package.GetName() + "." + e.GetName()
}

func (e *EnumGroup) GetImportPath() string {
	if e == nil {
		return ""
	}

	if e.ImportPath == "" && e.File != nil {
		return e.File.GetImportPath()
	}

	return e.ImportPath
}

func (e *EnumGroup) IsInternal(p *ProtoPackage) bool {
	if e == nil || p == nil {
		return false
	}

	return e.Package == p
}

type ProtoEnumField struct {
	*ProtoField[ProtoEnumField]
	*ConstField[ProtoEnumField, int32, int32]
	*OptionalField[ProtoEnumField]
}

func EnumField(name string, enum *EnumGroup) *ProtoEnumField {
	if enum == nil {
		log.Fatalf("Could not create the enum field %q because the enum given was nil.", name)
	}

	rules := make(map[string]any)
	options := make(map[string]any)

	ef := &ProtoEnumField{}
	internal := &protoFieldInternal{
		name:          name,
		goType:        "int32",
		protoType:     enum.GetName(),
		rules:         rules,
		protoBaseType: "enum",
		options:       options,
		enumRef:       enum,
		imports:       []string{enum.GetImportPath()},
	}

	ef.ProtoField = &ProtoField[ProtoEnumField]{
		protoFieldInternal: internal, self: ef,
	}
	ef.ConstField = &ConstField[ProtoEnumField, int32, int32]{constInternal: internal, self: ef}
	ef.OptionalField = &OptionalField[ProtoEnumField]{optionalInternal: internal, self: ef}

	return ef
}

func (ef *ProtoEnumField) Build(fieldNr uint32, imports Set) (FieldData, error) {
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

func (ef *ProtoEnumField) DefinedOnly() *ProtoEnumField {
	ef.rules["defined_only"] = true
	return ef
}
