package schemabuilder

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"strconv"
	"strings"
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

func (b *ProtoFieldExternal[T]) IgnoreIfUnpopulated() *T {
	b.options["(buf.validate.field).ignore"] = "IGNORE_IF_UNPOPULATED"
	return b.self
}

func (b *ProtoFieldExternal[T]) IgnoreIfDefaultValue() *T {
	b.options["(buf.validate.field).ignore"] = "IGNORE_IF_DEFAULT_VALUE"
	return b.self
}

func (b *ProtoFieldExternal[T]) IgnoreAlways() *T {
	b.options["(buf.validate.field).ignore"] = "IGNORE_ALWAYS"
	return b.self
}

func (b *ProtoFieldExternal[T]) Deprecated() *T {
	b.options["deprecated"] = "true"
	return b.self
}

func (b *ProtoFieldExternal[T]) CelField(o CelFieldOpts) *T {
	b.celOptions = append(b.celOptions, CelFieldOpts{
		Id: o.Id, Expression: o.Expression, Message: o.Message,
	})

	return b.self
}

func (b *ProtoFieldExternal[T]) Repeated() *T {
	if b.optional {
		b.errors = append(b.errors, fmt.Errorf("A field cannot be repeated and optional."))
	}
	b.repeated = true
	return b.self
}

// To refine with proto type and go type
func (b *ProtoFieldExternal[T]) Const(val any) *T {
	valType := reflect.TypeOf(val).String()
	if valType != b.fieldType {
		err := fmt.Errorf("The type for const does not match.\nField type: %s\nConst type: %s", b.fieldType, valType)
		b.errors = append(b.errors, err)
	}
	return b.self
}

func (b *ProtoFieldExternal[T]) Required() *T {
	if b.optional {
		b.errors = append(b.errors, fmt.Errorf("A field cannot be required and optional."))
	}
	b.options["(buf.validate.field).required"] = "true"
	b.required = true
	return b.self
}

func (b *ProtoFieldExternal[T]) Example(e string) *T {
	// Make this specific to the single validators
	b.options["(buf.validate.field).timestamp.example"] = e
	return b.self
}

type GenericField struct {
	*ProtoFieldExternal[GenericField]
}

func ImportedType(fieldNr int, name string, importPath string) *GenericField {
	imports := make(Set)
	options := make(map[string]string)
	imports[importPath] = present
	internal := &protoFieldInternal{fieldNr: fieldNr, fieldType: name, imports: imports, options: options}

	gf := &GenericField{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[GenericField]{
		protoFieldInternal: internal,
		self:               gf,
	}
	return gf
}

func ProtoTimestamp(fieldNr int) *GenericField {
	imports := make(Set)
	options := make(map[string]string)
	imports["google/protobuf/timestamp.proto"] = present
	internal := &protoFieldInternal{fieldNr: fieldNr, fieldType: "google.protobuf.Timestamp", imports: imports, options: options}

	gf := &GenericField{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[GenericField]{
		protoFieldInternal: internal,
		self:               gf,
	}
	return gf
}

func FieldMask(fieldNr int) *GenericField {
	imports := make(Set)
	options := make(map[string]string)
	imports["google/protobuf/field_mask.proto"] = present
	internal := &protoFieldInternal{fieldNr: fieldNr, fieldType: "fieldmask", imports: imports, options: options}

	gf := &GenericField{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[GenericField]{
		protoFieldInternal: internal,
		self:               gf,
	}
	return gf
}

func InternalType(fieldNr int, name string) *GenericField {
	imports := make(Set)
	options := make(map[string]string)
	internal := &protoFieldInternal{fieldNr: fieldNr, fieldType: name, imports: imports, options: options}

	gf := &GenericField{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[GenericField]{
		protoFieldInternal: internal,
		self:               gf,
	}
	return gf
}

type BytesField struct {
	*ProtoFieldExternal[BytesField]
	*LengthableField[BytesField]
}

func ProtoBytes(fieldNumber int) *BytesField {
	imports := make(Set)
	options := make(map[string]string)
	internal := &protoFieldInternal{fieldNr: fieldNumber, fieldType: "bytes", imports: imports, options: options}

	bf := &BytesField{}
	bf.ProtoFieldExternal = &ProtoFieldExternal[BytesField]{
		protoFieldInternal: internal,
		self:               bf,
	}
	bf.LengthableField = &LengthableField[BytesField]{
		internal:     internal,
		self:         bf,
		optionPrefix: "(buf.validate.field).bytes.",
	}
	return bf
}

type LengthableField[T any] struct {
	internal     *protoFieldInternal
	self         *T
	optionPrefix string
}

func (l *LengthableField[T]) MinLen(n int) *T {
	l.internal.options[l.optionPrefix+"min_len"] = strconv.Itoa(n)
	return l.self
}

func (l *LengthableField[T]) MaxLen(n int) *T {
	l.internal.options[l.optionPrefix+"max_len"] = strconv.Itoa(n)
	return l.self
}

func (l *LengthableField[T]) Len(n int) *T {
	l.internal.options[l.optionPrefix+"len"] = strconv.Itoa(n)
	return l.self
}

type StringField struct {
	*ProtoFieldExternal[StringField]
	*LengthableField[StringField]
}

func ProtoString(fieldNumber int) *StringField {
	imports := make(Set)
	options := make(map[string]string)
	internal := &protoFieldInternal{fieldNr: fieldNumber, fieldType: "string", imports: imports, options: options}

	sf := &StringField{}
	sf.ProtoFieldExternal = &ProtoFieldExternal[StringField]{
		protoFieldInternal: internal,
		self:               sf,
	}
	sf.LengthableField = &LengthableField[StringField]{
		internal:     internal,
		self:         sf,
		optionPrefix: "(buf.validate.field).string.",
	}
	return sf
}

type ProtoFieldExternal[T any] struct {
	*protoFieldInternal
	self *T
}

func (b *ProtoFieldExternal[T]) Optional() *T {
	if b.repeated {
		b.errors = append(b.errors, fmt.Errorf("A field cannot be repeated and optional."))
	}
	if b.required {
		b.errors = append(b.errors, fmt.Errorf("A field cannot be required and optional."))
	}
	b.optional = true
	return b.self
}
