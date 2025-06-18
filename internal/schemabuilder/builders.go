package schemabuilder

import (
	"errors"
	"fmt"
	"maps"
	"strings"

	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var present = struct{}{}

const indent = "  "
const indent2 = "    "

type Set map[string]struct{}

type Errors []error

type ProtoFieldData struct {
	Rules       map[string]any
	Options     []string
	ProtoType   string
	GoType      string
	Optional    bool
	FieldNr     int
	Name        string
	Imports     Set
	Deprecated  bool
	Repeated    bool
	Required    bool
	IsNonScalar bool
}

type protoFieldInternal struct {
	options         map[string]string
	rules           map[string]any
	repeatedOptions []string
	optional        bool
	fieldNr         int
	imports         Set
	protoType       string
	goType          string
	fieldMask       bool
	deprecated      bool
	errors          Errors
	required        bool
	isNonScalar     bool
}

type ProtoFieldBuilder interface {
	Build(fieldName string, imports Set) (ProtoFieldData, error)
}

func (b *protoFieldInternal) Build(fieldName string, imports Set) (ProtoFieldData, error) {
	if len(b.errors) > 0 {
		fieldErrors := strings.Builder{}
		for _, err := range b.errors {
			fieldErrors.WriteString(fmt.Sprintf("- %s\n", err.Error()))
		}

		return ProtoFieldData{}, errors.New(fieldErrors.String())
	}
	imports["buf/validate/validate.proto"] = present

	maps.Copy(imports, b.imports)

	options := GetOptions(b.options, b.repeatedOptions)

	return ProtoFieldData{Name: fieldName, Options: options, ProtoType: b.protoType, GoType: b.goType, Optional: b.optional, FieldNr: b.fieldNr, Rules: b.rules, IsNonScalar: b.isNonScalar}, nil
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) Options(o []ProtoOption) *BuilderT {
	for _, op := range o {
		b.options[op.Name] = op.Value
	}
	return b.self
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) IgnoreIfUnpopulated() *BuilderT {
	b.options["(buf.validate.field).ignore"] = "IGNORE_IF_UNPOPULATED"
	return b.self
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) IgnoreIfDefaultValue() *BuilderT {
	b.options["(buf.validate.field).ignore"] = "IGNORE_IF_DEFAULT_VALUE"
	return b.self
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) IgnoreAlways() *BuilderT {
	b.options["(buf.validate.field).ignore"] = "IGNORE_ALWAYS"
	return b.self
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) Deprecated() *BuilderT {
	b.options["deprecated"] = "true"
	return b.self
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) CelOption(o CelFieldOpts) *BuilderT {
	b.repeatedOptions = append(b.repeatedOptions, GetCelOption(CelFieldOpts{
		Id: o.Id, Expression: o.Expression, Message: o.Message,
	}))

	return b.self
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) Required() *BuilderT {
	if b.optional {
		b.errors = append(b.errors, fmt.Errorf("A field cannot be required and optional."))
	}
	b.options["(buf.validate.field).required"] = "true"
	b.required = true
	return b.self
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) Optional() *BuilderT {
	b.optional = true
	return b.self
}

type ProtoFieldExternal[BuilderT any, ValueT any] struct {
	*protoFieldInternal
	self *BuilderT
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) Const(val ValueT) *BuilderT {
	formattedVal, err := formatProtoValue(val)
	if err != nil {
		b.errors = append(b.errors, err)
		return b.self
	}

	// For duration and timestam pthis should not be the entire name but only the last part
	b.options[fmt.Sprintf("(buf.validate.field).%s.const", b.protoType)] = formattedVal
	return b.self
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) Example(val ValueT) *BuilderT {
	formattedVal, err := formatProtoValue(val)
	if err != nil {
		b.errors = append(b.errors, err)
		return b.self
	}

	// Make this repeatable
	b.repeatedOptions = append(b.repeatedOptions, fmt.Sprintf("(buf.validate.field).%s.example = %s", b.protoType, formattedVal))
	return b.self
}

type GenericField[ValueT any] struct {
	*ProtoFieldExternal[GenericField[ValueT], ValueT]
}

func ProtoTimestamp(fieldNr int) *GenericField[*timestamppb.Timestamp] {
	imports := make(Set)
	options := make(map[string]string)
	imports["google/protobuf/timestamp.proto"] = present
	internal := &protoFieldInternal{fieldNr: fieldNr, protoType: "google.protobuf.Timestamp", goType: "timestamp", imports: imports, options: options, isNonScalar: true}

	gf := &GenericField[*timestamppb.Timestamp]{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[GenericField[*timestamppb.Timestamp], *timestamppb.Timestamp]{
		protoFieldInternal: internal,
		self:               gf,
	}
	return gf
}

func FieldMask(fieldNr int) *GenericField[*fieldmaskpb.FieldMask] {
	imports := make(Set)
	options := make(map[string]string)
	imports["google/protobuf/field_mask.proto"] = present
	internal := &protoFieldInternal{fieldNr: fieldNr, protoType: "google.protobuf.FieldMask", goType: "fieldmask", imports: imports, options: options, isNonScalar: true}

	gf := &GenericField[*fieldmaskpb.FieldMask]{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[GenericField[*fieldmaskpb.FieldMask], *fieldmaskpb.FieldMask]{
		protoFieldInternal: internal,
		self:               gf,
	}
	return gf
}

type ProtoRepeatedBuilder struct {
	rules  map[string]any
	field  ProtoFieldBuilder
	unique bool
}

func RepeatedField(b ProtoFieldBuilder) *ProtoRepeatedBuilder {
	return &ProtoRepeatedBuilder{
		rules: map[string]any{}, field: b,
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

	if err != nil {
		return ProtoFieldData{}, err
	}

	if len(fieldData.Rules) > 0 {
		processedRules := 0
		stringRule := strings.Builder{}
		stringRule.WriteString("(buf.validate.field).repeated.items = {\n")
		stringRule.WriteString(fmt.Sprintf("  %s: {\n", fieldData.ProtoType))
		for name, value := range fieldData.Rules {
			if name == "required" {
				options = append(options, "(buf.validate.field).required = true")
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

	for rule, value := range b.rules {
		stringValue, fmtErr := formatProtoValue(value)
		if fmtErr != nil {
			err = fmt.Errorf("- %s\n%w", fmtErr, err)
		} else {
			options = append(options, fmt.Sprintf("%s = %s", rule, stringValue))
		}
	}

	return ProtoFieldData{Name: fieldName, ProtoType: fieldData.ProtoType, GoType: "[]" + fieldData.GoType, Optional: fieldData.Optional, FieldNr: fieldData.FieldNr, Repeated: true, Options: options, IsNonScalar: true}, nil
}

func (b *ProtoRepeatedBuilder) Unique() *ProtoRepeatedBuilder {
	b.unique = true
	return b
}
