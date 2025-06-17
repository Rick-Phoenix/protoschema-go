package schemabuilder

import (
	"errors"
	"fmt"
	"maps"
	"strconv"
	"strings"

	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var present = struct{}{}

const indent = "  "

type Set map[string]struct{}

type Errors []error

type ProtoFieldData struct {
	Options    []string
	ProtoType  string
	GoType     string
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
	protoType  string
	goType     string
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

	return ProtoFieldData{Name: fieldName, Options: options, ProtoType: b.protoType, GoType: b.goType, Optional: b.optional, FieldNr: b.fieldNr, Repeated: b.repeated}, nil
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
	l.internal.options[l.internal.protoType+"min_len"] = strconv.Itoa(n)
	return l.self
}

func (l *LengthableField[T]) MaxLen(n int) *T {
	l.internal.options[l.internal.protoType+"max_len"] = strconv.Itoa(n)
	return l.self
}

func (l *LengthableField[T]) Len(n int) *T {
	l.internal.options[l.internal.protoType+"len"] = strconv.Itoa(n)
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
	formattedVal, err := formatProtoConstValue(val, b.protoType)
	if err != nil {
		b.errors = append(b.errors, err)
		return b.self
	}

	b.options[fmt.Sprintf("(buf.validate.field).%s.const", b.protoType)] = formattedVal
	return b.self
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) Example(val ValueT) *BuilderT {
	formattedVal, err := formatProtoConstValue(val, b.protoType)
	if err != nil {
		b.errors = append(b.errors, err)
		return b.self
	}

	b.options[fmt.Sprintf("(buf.validate.field).%s.example", b.protoType)] = formattedVal
	return b.self
}

type IntField struct {
	*ProtoFieldExternal[IntField, int32]
}

func ProtoInt(fieldNumber int) *IntField {
	imports := make(Set)
	options := make(map[string]string)
	internal := &protoFieldInternal{fieldNr: fieldNumber, protoType: "int32", goType: "int32", imports: imports, options: options}

	integerField := &IntField{}
	integerField.ProtoFieldExternal = &ProtoFieldExternal[IntField, int32]{
		protoFieldInternal: internal,
		self:               integerField,
	}
	return integerField
}

type StringField struct {
	*ProtoFieldExternal[StringField, string]
	*LengthableField[StringField]
}

func ProtoString(fieldNumber int) *StringField {
	imports := make(Set)
	options := make(map[string]string)
	internal := &protoFieldInternal{fieldNr: fieldNumber, protoType: "string", goType: "string", imports: imports, options: options}

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
	*ProtoFieldExternal[BytesField, []byte]
	*LengthableField[BytesField]
}

func ProtoBytes(fieldNumber int) *BytesField {
	imports := make(Set)
	options := make(map[string]string)
	internal := &protoFieldInternal{fieldNr: fieldNumber, protoType: "bytes", goType: "bytes", imports: imports, options: options}

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
	internal := &protoFieldInternal{fieldNr: fieldNr, protoType: name, goType: "any", imports: imports, options: options}

	gf := &GenericField[any]{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[GenericField[any], any]{
		protoFieldInternal: internal,
		self:               gf,
	}
	return gf
}

func ProtoTimestamp(fieldNr int) *GenericField[*timestamppb.Timestamp] {
	imports := make(Set)
	options := make(map[string]string)
	imports["google/protobuf/timestamp.proto"] = present
	internal := &protoFieldInternal{fieldNr: fieldNr, protoType: "google.protobuf.Timestamp", goType: "timestamp", imports: imports, options: options}

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
	internal := &protoFieldInternal{fieldNr: fieldNr, protoType: "google.protobuf.FieldMask", goType: "fieldmask", imports: imports, options: options}

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
	internal := &protoFieldInternal{fieldNr: fieldNr, protoType: name, goType: "any", imports: imports, options: options}

	gf := &GenericField[any]{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[GenericField[any], any]{
		protoFieldInternal: internal,
		self:               gf,
	}
	return gf
}
