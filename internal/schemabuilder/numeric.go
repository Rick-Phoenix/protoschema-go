package schemabuilder

import (
	"fmt"
)

type Number interface {
	uint | int8 | int16 | int32 | int64 |
		uint8 | uint16 | uint32 | uint64 |
		float32 | float64
}

type NumericField[BuilderT any, ValueT comparable] struct {
	*protoFieldInternal
	*FieldWithConst[BuilderT, ValueT, ValueT]
	self *BuilderT

	hasLtOrLte bool
	hasGtOrGte bool

	isFloatType bool
}

func newNumericField[BuilderT any, ValueT comparable](pfi *protoFieldInternal, self *BuilderT, isFloat bool) *NumericField[BuilderT, ValueT] {
	return &NumericField[BuilderT, ValueT]{
		protoFieldInternal: pfi,
		self:               self,
		isFloatType:        isFloat,
		FieldWithConst: &FieldWithConst[BuilderT, ValueT, ValueT]{
			internal: pfi,
			self:     self,
		},
	}
}

func (nf *NumericField[BuilderT, ValueT]) Lt(val ValueT) *BuilderT {
	if nf.hasLtOrLte {
		// Check if larger than gt
		nf.errors = append(nf.errors, fmt.Errorf("A numeric field cannot have both 'lt' and 'lte' rules."))
	}
	nf.rules["lt"] = val
	nf.hasLtOrLte = true
	return nf.self
}

func (nf *NumericField[BuilderT, ValueT]) Lte(val ValueT) *BuilderT {
	if nf.hasLtOrLte {
		nf.errors = append(nf.errors, fmt.Errorf("A numeric field cannot have both 'lt' and 'lte' rules."))
	}
	nf.rules["lte"] = val
	nf.hasLtOrLte = true
	return nf.self
}

func (nf *NumericField[BuilderT, ValueT]) Gt(val ValueT) *BuilderT {
	if nf.hasGtOrGte {
		nf.errors = append(nf.errors, fmt.Errorf("A numeric field cannot have both 'gt' and 'gte' rules."))
	}
	nf.rules["gt"] = val
	nf.hasGtOrGte = true
	return nf.self
}

func (nf *NumericField[BuilderT, ValueT]) Gte(val ValueT) *BuilderT {
	if nf.hasGtOrGte {
		nf.errors = append(nf.errors, fmt.Errorf("A numeric field cannot have both 'gt' and 'gte' rules."))
	}
	nf.rules["gte"] = val
	nf.hasGtOrGte = true
	return nf.self
}

func (nf *NumericField[BuilderT, ValueT]) Finite() *BuilderT {
	if !nf.isFloatType {
		nf.errors = append(nf.errors, fmt.Errorf("'finite' rule is only applicable to float and double types."))
	}
	nf.rules["finite"] = true
	return nf.self
}

type IntField struct {
	*ProtoFieldExternal[IntField, int32]
	*NumericField[IntField, int32]
}

func ProtoInt32(fieldNumber uint) *IntField {
	options := make(map[string]string)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		fieldNr:   fieldNumber,
		protoType: "int32",
		goType:    "int32",
		options:   options,
		rules:     rules,
	}

	integerField := &IntField{}
	integerField.ProtoFieldExternal = &ProtoFieldExternal[IntField, int32]{
		protoFieldInternal: internal,
		self:               integerField,
	}
	integerField.NumericField = newNumericField[IntField, int32](internal, integerField, false)
	return integerField
}

type FloatField struct {
	*ProtoFieldExternal[FloatField, float32]
	*NumericField[FloatField, float32]
}

func ProtoFloat(fieldNumber uint) *FloatField {
	options := make(map[string]string)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		fieldNr:   fieldNumber,
		protoType: "float",
		goType:    "float32",
		options:   options,
		rules:     rules,
	}

	floatField := &FloatField{}
	floatField.ProtoFieldExternal = &ProtoFieldExternal[FloatField, float32]{
		protoFieldInternal: internal,
		self:               floatField,
	}
	floatField.NumericField = newNumericField[FloatField, float32](internal, floatField, true)
	return floatField
}

type DoubleField struct {
	*ProtoFieldExternal[DoubleField, float64]
	*NumericField[DoubleField, float64]
}

func ProtoDouble(fieldNumber uint) *DoubleField {
	options := make(map[string]string)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		fieldNr:   fieldNumber,
		protoType: "double",
		goType:    "float64",
		options:   options,
		rules:     rules,
	}

	floatField := &DoubleField{}
	floatField.ProtoFieldExternal = &ProtoFieldExternal[DoubleField, float64]{
		protoFieldInternal: internal,
		self:               floatField,
	}
	floatField.NumericField = newNumericField[DoubleField, float64](internal, floatField, true)
	return floatField
}

type Int64Field struct {
	*ProtoFieldExternal[Int64Field, int64]
	*NumericField[Int64Field, int64]
}

func ProtoInt64(fieldNumber uint) *Int64Field {
	options := make(map[string]string)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		fieldNr:   fieldNumber,
		protoType: "int64",
		goType:    "int64",
		options:   options,
		rules:     rules,
	}

	int64Field := &Int64Field{}
	int64Field.ProtoFieldExternal = &ProtoFieldExternal[Int64Field, int64]{
		protoFieldInternal: internal,
		self:               int64Field,
	}
	int64Field.NumericField = newNumericField[Int64Field, int64](internal, int64Field, false)
	return int64Field
}

type UInt32Field struct {
	*ProtoFieldExternal[UInt32Field, uint32]
	*NumericField[UInt32Field, uint32]
}

func ProtoUInt32(fieldNumber uint) *UInt32Field {
	options := make(map[string]string)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		fieldNr:   fieldNumber,
		protoType: "uint32",
		goType:    "uint32",
		options:   options,
		rules:     rules,
	}

	uint32Field := &UInt32Field{}
	uint32Field.ProtoFieldExternal = &ProtoFieldExternal[UInt32Field, uint32]{
		protoFieldInternal: internal,
		self:               uint32Field,
	}
	uint32Field.NumericField = newNumericField[UInt32Field, uint32](internal, uint32Field, false)
	return uint32Field
}

type UInt64Field struct {
	*ProtoFieldExternal[UInt64Field, uint64]
	*NumericField[UInt64Field, uint64]
}

func ProtoUInt64(fieldNumber uint) *UInt64Field {
	options := make(map[string]string)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		fieldNr:   fieldNumber,
		protoType: "uint64",
		goType:    "uint64",
		options:   options,
		rules:     rules,
	}

	uint64Field := &UInt64Field{}
	uint64Field.ProtoFieldExternal = &ProtoFieldExternal[UInt64Field, uint64]{
		protoFieldInternal: internal,
		self:               uint64Field,
	}
	uint64Field.NumericField = newNumericField[UInt64Field, uint64](internal, uint64Field, false)
	return uint64Field
}

type SInt32Field struct {
	*ProtoFieldExternal[SInt32Field, int32]
	*NumericField[SInt32Field, int32]
}

func ProtoSInt32(fieldNumber uint) *SInt32Field {
	options := make(map[string]string)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		fieldNr:   fieldNumber,
		protoType: "sint32",
		goType:    "int32",
		options:   options,
		rules:     rules,
	}

	sint32Field := &SInt32Field{}
	sint32Field.ProtoFieldExternal = &ProtoFieldExternal[SInt32Field, int32]{
		protoFieldInternal: internal,
		self:               sint32Field,
	}
	sint32Field.NumericField = newNumericField[SInt32Field, int32](internal, sint32Field, false)
	return sint32Field
}

type SInt64Field struct {
	*ProtoFieldExternal[SInt64Field, int64]
	*NumericField[SInt64Field, int64]
}

func ProtoSInt64(fieldNumber uint) *SInt64Field {
	options := make(map[string]string)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		fieldNr:   fieldNumber,
		protoType: "sint64",
		goType:    "int64",
		options:   options,
		rules:     rules,
	}

	sint64Field := &SInt64Field{}
	sint64Field.ProtoFieldExternal = &ProtoFieldExternal[SInt64Field, int64]{
		protoFieldInternal: internal,
		self:               sint64Field,
	}
	sint64Field.NumericField = newNumericField[SInt64Field, int64](internal, sint64Field, false)
	return sint64Field
}

type Fixed32Field struct {
	*ProtoFieldExternal[Fixed32Field, uint32]
	*NumericField[Fixed32Field, uint32]
}

func ProtoFixed32(fieldNumber uint) *Fixed32Field {
	options := make(map[string]string)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		fieldNr:   fieldNumber,
		protoType: "fixed32",
		goType:    "uint32",
		options:   options,
		rules:     rules,
	}

	fixed32Field := &Fixed32Field{}
	fixed32Field.ProtoFieldExternal = &ProtoFieldExternal[Fixed32Field, uint32]{
		protoFieldInternal: internal,
		self:               fixed32Field,
	}
	fixed32Field.NumericField = newNumericField[Fixed32Field, uint32](internal, fixed32Field, false)
	return fixed32Field
}

type Fixed64Field struct {
	*ProtoFieldExternal[Fixed64Field, uint64]
	*NumericField[Fixed64Field, uint64]
}

func ProtoFixed64(fieldNumber uint) *Fixed64Field {
	options := make(map[string]string)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		fieldNr:   fieldNumber,
		protoType: "fixed64",
		goType:    "uint64",
		options:   options,
		rules:     rules,
	}

	fixed64Field := &Fixed64Field{}
	fixed64Field.ProtoFieldExternal = &ProtoFieldExternal[Fixed64Field, uint64]{
		protoFieldInternal: internal,
		self:               fixed64Field,
	}
	fixed64Field.NumericField = newNumericField[Fixed64Field, uint64](internal, fixed64Field, false)
	return fixed64Field
}

type SFixed32Field struct {
	*ProtoFieldExternal[SFixed32Field, int32]
	*NumericField[SFixed32Field, int32]
}

func ProtoSFixed32(fieldNumber uint) *SFixed32Field {
	options := make(map[string]string)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		fieldNr:   fieldNumber,
		protoType: "sfixed32",
		goType:    "int32",
		options:   options,
		rules:     rules,
	}

	sfixed32Field := &SFixed32Field{}
	sfixed32Field.ProtoFieldExternal = &ProtoFieldExternal[SFixed32Field, int32]{
		protoFieldInternal: internal,
		self:               sfixed32Field,
	}
	sfixed32Field.NumericField = newNumericField[SFixed32Field, int32](internal, sfixed32Field, false)
	return sfixed32Field
}

type SFixed64Field struct {
	*ProtoFieldExternal[SFixed64Field, int64]
	*NumericField[SFixed64Field, int64]
}

func ProtoSFixed64(fieldNumber uint) *SFixed64Field {
	options := make(map[string]string)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		fieldNr:   fieldNumber,
		protoType: "sfixed64",
		goType:    "int64",
		options:   options,
		rules:     rules,
	}

	sfixed64Field := &SFixed64Field{}
	sfixed64Field.ProtoFieldExternal = &ProtoFieldExternal[SFixed64Field, int64]{
		protoFieldInternal: internal,
		self:               sfixed64Field,
	}
	sfixed64Field.NumericField = newNumericField[SFixed64Field, int64](internal, sfixed64Field, false)
	return sfixed64Field
}
