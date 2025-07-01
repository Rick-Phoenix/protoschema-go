package schemabuilder

import (
	"errors"
	"maps"

	"github.com/labstack/gommon/log"
)

// The members of an enum group.
type EnumMembers map[int32]string

// The schema for a protobuf Enum.This should be created with the constructor from the FileSchema or MessageSchema instances to automatically populate the Package, File and Message fields. It can also be used as a struct to define an Enum that was not defined by using this library.
type EnumGroup struct {
	// The enum's name. If this enum was defined in a message, the GetName method will automatically prepend the parent message's name.
	Name string
	// The members of this enum group.
	Members         EnumMembers
	ReservedNames   []string
	ReservedNumbers []int32
	ReservedRanges  []Range
	// Custom options for this enum. A preset for the allow_alias option is available in this package under Options.AllowAlias.
	Options []ProtoOption
	// The package that this enum belongs to. Automatically set when using the constructors.
	Package *ProtoPackage
	// The file that this enum belongs to. Automatically set when using the constructors.
	File *FileSchema
	// The message that this enum belongs to (if defined within a message). Automatically set when using the constructors.
	Message *MessageSchema
	// Custom metadata to use with the FileHook, which will receive the data for the EnumGroups in it, including this map.
	Metadata map[string]any
	// The import path to this enum's file. Automatically set when using the constructors.
	ImportPath string
}

// Returns the enum's name, prepending the name of the parent message (and its own parent messages), if there is one.
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

// If the argument package is the same as this enum's, it will return the enum's name (with the parent message's name, if there is one). Otherwise, it will return the full name, including the package that it belongs to.
func (e *EnumGroup) GetFullName(p *ProtoPackage) string {
	if e == nil {
		return ""
	}

	if e.Package == nil || e.Package == p {
		return e.GetName()
	}

	if e.Message != nil {
		return e.Message.GetFullName(p) + "." + e.Name
	}

	return e.Package.GetName() + "." + e.GetName()
}

// Returns the import path of the file that this enum belongs to (if one is defined).
func (e *EnumGroup) GetImportPath() string {
	if e == nil {
		return ""
	}

	if e.ImportPath == "" && e.File != nil {
		return e.File.GetImportPath()
	}

	return e.ImportPath
}

// Returns true if the argument package is the same as this enum's.
func (e *EnumGroup) IsInternal(p *ProtoPackage) bool {
	if e == nil || p == nil {
		return false
	}

	return e.Package == p
}

// A message field with an enum type.
type ProtoEnumField struct {
	*ProtoField[ProtoEnumField]
	*ConstField[ProtoEnumField, int32, int32]
	*OptionalField[ProtoEnumField]
}

// The constructor for an enum field.
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

// The method that processes the field's schema and returns its data. Used to satisfy the FieldBuilder interface. Mostly for internal use.
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

// Rule: this field must contain one of the defined values for its enum type.
func (ef *ProtoEnumField) DefinedOnly() *ProtoEnumField {
	ef.rules["defined_only"] = true
	return ef
}
