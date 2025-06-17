package schemabuilder

import (
	"encoding/base64"
	"errors"
	"fmt"
	"maps"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var present = struct{}{}

const indent = "  "

type Set map[string]struct{}

type Errors []error

type ProtoFieldData struct {
	Options    []string
	FieldType  string
	Optional   bool
	FieldNr    int
	Name       string
	Imports    Set
	Deprecated bool
	Repeated   bool
}

type protoFieldInternal struct {
	options    map[string]string
	celOptions []CelFieldOpts
	optional   bool
	fieldNr    int
	imports    Set
	fieldType  string
	fieldMask  bool
	deprecated bool
	repeated   bool
	errors     Errors
	required   bool
}

type ProtoFieldBuilder interface {
	Build(fieldName string, imports Set) (ProtoFieldData, error)
}

func (b *protoFieldInternal) Build(fieldName string, imports Set) (ProtoFieldData, error) {
	if len(b.errors) > 0 {
		fieldErrors := strings.Builder{}
		fieldErrors.WriteString(fmt.Sprintf("Errors for field %s:\n", fieldName))
		for _, err := range b.errors {
			fieldErrors.WriteString(fmt.Sprintf("%s- %s\n", indent, err.Error()))
		}

		return ProtoFieldData{}, errors.New(fieldErrors.String())
	}
	imports["buf/validate/validate.proto"] = present

	maps.Copy(imports, b.imports)

	options := GetOptions(b.options, b.celOptions)

	return ProtoFieldData{Name: fieldName, Options: options, FieldType: b.fieldType, Optional: b.optional, FieldNr: b.fieldNr, Repeated: b.repeated}, nil
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

func (b *ProtoFieldExternal[BuilderT, ValueT]) CelField(o CelFieldOpts) *BuilderT {
	b.celOptions = append(b.celOptions, CelFieldOpts{
		Id: o.Id, Expression: o.Expression, Message: o.Message,
	})

	return b.self
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) Repeated() *BuilderT {
	if b.optional {
		b.errors = append(b.errors, fmt.Errorf("A field cannot be repeated and optional."))
	}
	b.repeated = true
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

type LengthableField[T any] struct {
	internal *protoFieldInternal
	self     *T
}

func (l *LengthableField[T]) MinLen(n int) *T {
	l.internal.options[l.internal.fieldType+"min_len"] = strconv.Itoa(n)
	return l.self
}

func (l *LengthableField[T]) MaxLen(n int) *T {
	l.internal.options[l.internal.fieldType+"max_len"] = strconv.Itoa(n)
	return l.self
}

func (l *LengthableField[T]) Len(n int) *T {
	l.internal.options[l.internal.fieldType+"len"] = strconv.Itoa(n)
	return l.self
}

type ProtoFieldExternal[BuilderT any, ValueT any] struct {
	*protoFieldInternal
	self *BuilderT
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) Optional() *BuilderT {
	if b.repeated {
		b.errors = append(b.errors, fmt.Errorf("A field cannot be repeated and optional."))
	}
	if b.required {
		b.errors = append(b.errors, fmt.Errorf("A field cannot be required and optional."))
	}
	b.optional = true
	return b.self
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) Const(val ValueT) *BuilderT {
	// Map the internal fieldType (e.g., "google.protobuf.Timestamp")
	// to the short name used in protovalidate options (e.g., "timestamp").
	protoTypeName := b.protoFieldInternal.getProtoOptionTypeName()
	if protoTypeName == "" {
		b.errors = append(b.errors, fmt.Errorf("const validation not supported for proto field type %q", b.protoFieldInternal.fieldType))
		return b.self
	}

	// Format the Go value into a string suitable for the proto option.
	formattedVal, err := formatProtoConstValue(val, protoTypeName)
	if err != nil {
		b.errors = append(b.errors, err)
		return b.self
	}

	b.options[fmt.Sprintf("(buf.validate.field).%s.const", protoTypeName)] = formattedVal
	return b.self
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) Example(val ValueT) *BuilderT {
	// Map the internal fieldType (e.g., "google.protobuf.Timestamp")
	// to the short name used in protovalidate options (e.g., "timestamp").
	protoTypeName := b.protoFieldInternal.getProtoOptionTypeName()
	if protoTypeName == "" {
		b.errors = append(b.errors, fmt.Errorf("example not supported for proto field type %q", b.protoFieldInternal.fieldType))
		return b.self
	}

	formattedVal, err := formatProtoConstValue(val, protoTypeName)
	if err != nil {
		b.errors = append(b.errors, err)
		return b.self
	}

	b.options[fmt.Sprintf("(buf.validate.field).%s.example", protoTypeName)] = formattedVal
	return b.self
}

func (pfi *protoFieldInternal) getProtoOptionTypeName() string {
	switch pfi.fieldType {
	case "string":
		return "string"
	case "int32", "sint32", "fixed32":
		return "int32"
	case "int64", "sint64", "fixed64":
		return "int64"
	case "uint32":
		return "uint32"
	case "uint64":
		return "uint64"
	case "bool":
		return "bool"
	case "double":
		return "double"
	case "float":
		return "float"
	case "bytes":
		return "bytes"
	case "google.protobuf.Timestamp":
		return "timestamp"
	case "google.protobuf.Duration":
		return "duration"
	case "google.protobuf.FieldMask":
		return "field_mask"
	// Add more mappings for other scalar or well-known types as needed
	default:
		// For custom message types or enums, 'const' might not be directly applicable
		// or requires a different approach/option.
		return ""
	}
}

// formatProtoConstValue formats a Go value into a string for proto options.
// This function centralizes the logic for how different Go types should be represented
// as string literals in the .proto file for const validation.
func formatProtoConstValue(val any, protoTypeName string) (string, error) {
	switch protoTypeName {
	case "string":
		strVal, ok := val.(string)
		if !ok {
			return "", fmt.Errorf("expected string for %s const, got %T", protoTypeName, val)
		}
		return strconv.Quote(strVal), nil // Strings need to be quoted in proto options
	case "int32", "int64", "uint32", "uint64", "double", "float", "bool":
		return fmt.Sprintf("%v", val), nil // Scalar numbers and booleans don't need quotes
	case "bytes":
		byteVal, ok := val.([]byte)
		if !ok {
			return "", fmt.Errorf("expected []byte for %s const, got %T", protoTypeName, val)
		}
		// Protobuf bytes constants in string form are usually base64 encoded.
		return strconv.Quote(base64.StdEncoding.EncodeToString(byteVal)), nil
	case "timestamp":
		tsVal, ok := val.(*timestamppb.Timestamp)
		if !ok {
			return "", fmt.Errorf("expected *timestamppb.Timestamp for %s const, got %T", protoTypeName, val)
		}
		// protovalidate typically expects timestamp constants in RFC3339 format or similar for simple const.
		// If `protovalidate` expects a timestamp literal like `{seconds: 123, nanos: 456}`,
		// this formatting would need to return that specific literal string.
		return strconv.Quote(tsVal.AsTime().Format(time.RFC3339Nano)), nil // RFC3339 with nanoseconds and quotes
	case "duration":
		durVal, ok := val.(*durationpb.Duration)
		if !ok {
			return "", fmt.Errorf("expected *durationpb.Duration for %s const, got %T", protoTypeName, val)
		}
		// Assuming standard Go duration string format (e.g., "1h30m") quoted.
		return strconv.Quote(durVal.AsDuration().String()), nil
	default:
		return "", fmt.Errorf("unsupported Go type %T for %s const validation", val, protoTypeName)
	}
}

type IntField struct {
	*ProtoFieldExternal[IntField, int32] // ValueT is 'int32'
}

func ProtoInt(fieldNumber int) *IntField {
	imports := make(Set)
	options := make(map[string]string)
	internal := &protoFieldInternal{fieldNr: fieldNumber, fieldType: "int32", imports: imports, options: options} // Use "int32" as proto field type

	integerField := &IntField{}
	integerField.ProtoFieldExternal = &ProtoFieldExternal[IntField, int32]{ // Specify 'int32' for ValueT
		protoFieldInternal: internal,
		self:               integerField,
	}
	return integerField
}

type StringField struct {
	*ProtoFieldExternal[StringField, string] // ValueT is 'string'
	*LengthableField[StringField]
}

func ProtoString(fieldNumber int) *StringField {
	imports := make(Set)
	options := make(map[string]string)
	internal := &protoFieldInternal{fieldNr: fieldNumber, fieldType: "string", imports: imports, options: options}

	sf := &StringField{}
	sf.ProtoFieldExternal = &ProtoFieldExternal[StringField, string]{
		protoFieldInternal: internal,
		self:               sf,
	}
	sf.LengthableField = &LengthableField[StringField]{
		internal: internal,
		self:     sf,
	}
	return sf
}

type BytesField struct {
	*ProtoFieldExternal[BytesField, []byte] // ValueT is '[]byte'
	*LengthableField[BytesField]
}

func ProtoBytes(fieldNumber int) *BytesField {
	imports := make(Set)
	options := make(map[string]string)
	internal := &protoFieldInternal{fieldNr: fieldNumber, fieldType: "bytes", imports: imports, options: options}

	bf := &BytesField{}
	bf.ProtoFieldExternal = &ProtoFieldExternal[BytesField, []byte]{ // Specify '[]byte' for ValueT
		protoFieldInternal: internal,
		self:               bf,
	}
	bf.LengthableField = &LengthableField[BytesField]{
		internal: internal,
		self:     bf,
	}
	return bf
}

type GenericField[ValueT any] struct {
	*ProtoFieldExternal[GenericField[ValueT], ValueT]
}

func ImportedType(fieldNr int, name string, importPath string) *GenericField[any] {
	imports := make(Set)
	options := make(map[string]string)
	imports[importPath] = present
	internal := &protoFieldInternal{fieldNr: fieldNr, fieldType: name, imports: imports, options: options}

	gf := &GenericField[any]{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[GenericField[any], any]{
		protoFieldInternal: internal,
		self:               gf,
	}
	return gf
}

func ProtoTimestamp(fieldNr int) *GenericField[*timestamppb.Timestamp] { // ValueT is *timestamppb.Timestamp
	imports := make(Set)
	options := make(map[string]string)
	imports["google/protobuf/timestamp.proto"] = present
	internal := &protoFieldInternal{fieldNr: fieldNr, fieldType: "google.protobuf.Timestamp", imports: imports, options: options}

	gf := &GenericField[*timestamppb.Timestamp]{} // Instantiate with specific ValueT
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
	internal := &protoFieldInternal{fieldNr: fieldNr, fieldType: "google.protobuf.FieldMask", imports: imports, options: options} // Corrected proto field name

	gf := &GenericField[*fieldmaskpb.FieldMask]{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[GenericField[*fieldmaskpb.FieldMask], *fieldmaskpb.FieldMask]{
		protoFieldInternal: internal,
		self:               gf,
	}
	return gf
}

func InternalType(fieldNr int, name string) *GenericField[any] {
	imports := make(Set)
	options := make(map[string]string)
	internal := &protoFieldInternal{fieldNr: fieldNr, fieldType: name, imports: imports, options: options}

	gf := &GenericField[any]{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[GenericField[any], any]{
		protoFieldInternal: internal,
		self:               gf,
	}
	return gf
}
