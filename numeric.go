package schemabuilder

import (
	"cmp"
	"errors"
	"fmt"
)

// The generic numeric field struct, implemented by the various numeric field types.
type NumericField[BuilderT any, ValueT cmp.Ordered] struct {
	*ProtoField[BuilderT]
	*ConstField[BuilderT, ValueT, ValueT]
	*OptionalField[BuilderT]
	self *BuilderT

	hasLtOrLte bool
	hasGtOrGte bool

	lt  *ValueT
	lte *ValueT
	gt  *ValueT
	gte *ValueT

	isFloatType bool
}

func newNumericField[BuilderT any, ValueT cmp.Ordered](pfi *protoFieldInternal, self *BuilderT, isFloat bool) *NumericField[BuilderT, ValueT] {
	return &NumericField[BuilderT, ValueT]{
		ProtoField: &ProtoField[BuilderT]{
			protoFieldInternal: pfi,
			self:               self,
		},
		isFloatType: isFloat,
		ConstField: &ConstField[BuilderT, ValueT, ValueT]{
			constInternal: pfi,
			self:          self,
		},
		OptionalField: &OptionalField[BuilderT]{
			optionalInternal: pfi,
			self:             self,
		},
	}
}

// Rule: this numeric field must be smaller than the indicated value.
func (nf *NumericField[BuilderT, ValueT]) Lt(val ValueT) *BuilderT {
	if nf.hasLtOrLte {
		nf.errors = errors.Join(nf.errors, fmt.Errorf("A numeric field cannot have both 'lt' and 'lte' rules."))
	}

	if nf.gt != nil && *nf.gt >= val {
		nf.errors = errors.Join(nf.errors, fmt.Errorf("'gt' cannot be larger than or equal to 'lt'."))
	}
	if nf.gte != nil && *nf.gte >= val {
		nf.errors = errors.Join(nf.errors, fmt.Errorf("'gte' cannot be larger than or equal to 'lt'."))
	}
	nf.rules["lt"] = val
	nf.hasLtOrLte = true
	nf.lt = &val
	return nf.ProtoField.self
}

// Rule: this numeric field must be smaller than or equal to the indicated value.
func (nf *NumericField[BuilderT, ValueT]) Lte(val ValueT) *BuilderT {
	if nf.hasLtOrLte {
		nf.errors = errors.Join(nf.errors, fmt.Errorf("A numeric field cannot have both 'lt' and 'lte' rules."))
	}

	if nf.gt != nil && *nf.gt >= val {
		nf.errors = errors.Join(nf.errors, fmt.Errorf("'gt' cannot be larger than or equal to 'lte'."))
	}
	if nf.gte != nil && *nf.gte > val {
		nf.errors = errors.Join(nf.errors, fmt.Errorf("'gt' cannot be larger than 'lte'."))
	}
	nf.rules["lte"] = val
	nf.hasLtOrLte = true
	nf.lte = &val
	return nf.ProtoField.self
}

// Rule: this numeric field must be larger than the indicated value.
func (nf *NumericField[BuilderT, ValueT]) Gt(val ValueT) *BuilderT {
	if nf.hasGtOrGte {
		nf.errors = errors.Join(nf.errors, fmt.Errorf("A numeric field cannot have both 'gt' and 'gte' rules."))
	}

	if nf.lt != nil && *nf.lt <= val {
		nf.errors = errors.Join(nf.errors, fmt.Errorf("'lt' cannot be smaller than or equal to 'gt'."))
	}
	if nf.lte != nil && *nf.lte <= val {
		nf.errors = errors.Join(nf.errors, fmt.Errorf("'lte' cannot be smaller than or equal to 'gt'."))
	}
	nf.rules["gt"] = val
	nf.hasGtOrGte = true
	nf.gt = &val
	return nf.ProtoField.self
}

// Rule: this numeric field must be larger than or equal to the indicated value.
func (nf *NumericField[BuilderT, ValueT]) Gte(val ValueT) *BuilderT {
	if nf.hasGtOrGte {
		nf.errors = errors.Join(nf.errors, fmt.Errorf("A numeric field cannot have both 'gt' and 'gte' rules."))
	}

	if nf.lt != nil && *nf.lt <= val {
		nf.errors = errors.Join(nf.errors, fmt.Errorf("'lt' cannot be smaller than or equal to 'gte'."))
	}
	if nf.lte != nil && *nf.lte < val {
		nf.errors = errors.Join(nf.errors, fmt.Errorf("'lte' cannot be smaller than 'gte'."))
	}
	nf.rules["gte"] = val
	nf.hasGtOrGte = true
	nf.gte = &val
	return nf.ProtoField.self
}

// Rule: this numeric field must be finite. Only applicable to float and double field types.
func (nf *NumericField[BuilderT, ValueT]) Finite() *BuilderT {
	if !nf.isFloatType {
		nf.errors = errors.Join(nf.errors, fmt.Errorf("The 'finite' rule is only applicable to float and double types."))
	}
	nf.rules["finite"] = true
	return nf.ProtoField.self
}

// An instance of an int32 protobuf field.
type Int32Field struct {
	*NumericField[Int32Field, int32]
}

// The constructor for an int32 protobuf field.
func Int32(name string) *Int32Field {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "int32",
		goType:    "int32",
		options:   options,
		rules:     rules,
	}

	integerField := &Int32Field{}

	integerField.NumericField = newNumericField[Int32Field, int32](internal, integerField, false)
	return integerField
}

// An instance of a float protobuf field.
type FloatField struct {
	*NumericField[FloatField, float32]
}

// The constructor for a float protobuf field.
func Float(name string) *FloatField {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "float",
		goType:    "float32",
		options:   options,
		rules:     rules,
	}

	floatField := &FloatField{}
	floatField.ProtoField = &ProtoField[FloatField]{
		protoFieldInternal: internal,
		self:               floatField,
	}
	floatField.NumericField = newNumericField[FloatField, float32](internal, floatField, true)
	return floatField
}

// An instance of a double protobuf field.
type DoubleField struct {
	*NumericField[DoubleField, float64]
}

// The constructor for a double protobuf field.
func Double(name string) *DoubleField {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "double",
		goType:    "float64",
		options:   options,
		rules:     rules,
	}

	floatField := &DoubleField{}
	floatField.ProtoField = &ProtoField[DoubleField]{
		protoFieldInternal: internal,
		self:               floatField,
	}
	floatField.NumericField = newNumericField[DoubleField, float64](internal, floatField, true)
	return floatField
}

// An instance of an int64 protobuf field.
type Int64Field struct {
	*NumericField[Int64Field, int64]
}

// The constructor for an int64 protobuf field.
func Int64(name string) *Int64Field {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "int64",
		goType:    "int64",
		options:   options,
		rules:     rules,
	}

	int64Field := &Int64Field{}
	int64Field.ProtoField = &ProtoField[Int64Field]{
		protoFieldInternal: internal,
		self:               int64Field,
	}
	int64Field.NumericField = newNumericField[Int64Field, int64](internal, int64Field, false)
	return int64Field
}

// An instance of a uint32 protobuf field.
type UInt32Field struct {
	*NumericField[UInt32Field, uint32]
}

// The constructor for a uint32 protobuf field.
func UInt32(name string) *UInt32Field {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "uint32",
		goType:    "uint32",
		options:   options,
		rules:     rules,
	}

	uint32Field := &UInt32Field{}
	uint32Field.ProtoField = &ProtoField[UInt32Field]{
		protoFieldInternal: internal,
		self:               uint32Field,
	}
	uint32Field.NumericField = newNumericField[UInt32Field, uint32](internal, uint32Field, false)
	return uint32Field
}

// An instance of a uint64 protobuf field.
type UInt64Field struct {
	*NumericField[UInt64Field, uint64]
}

// The constructor for a uint64 protobuf field.
func UInt64(name string) *UInt64Field {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "uint64",
		goType:    "uint64",
		options:   options,
		rules:     rules,
	}

	uint64Field := &UInt64Field{}
	uint64Field.ProtoField = &ProtoField[UInt64Field]{
		protoFieldInternal: internal,
		self:               uint64Field,
	}
	uint64Field.NumericField = newNumericField[UInt64Field, uint64](internal, uint64Field, false)
	return uint64Field
}

// An instance of a sint32 protobuf field.
type SInt32Field struct {
	*NumericField[SInt32Field, int32]
}

// The constructor for a sint32 protobuf field.
func SInt32(name string) *SInt32Field {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "sint32",
		goType:    "int32",
		options:   options,
		rules:     rules,
	}

	sint32Field := &SInt32Field{}
	sint32Field.ProtoField = &ProtoField[SInt32Field]{
		protoFieldInternal: internal,
		self:               sint32Field,
	}
	sint32Field.NumericField = newNumericField[SInt32Field, int32](internal, sint32Field, false)
	return sint32Field
}

// An instance of a sint64 protobuf field.
type SInt64Field struct {
	*NumericField[SInt64Field, int64]
}

// The constructor for a sint64 protobuf field.
func SInt64(name string) *SInt64Field {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "sint64",
		goType:    "int64",
		options:   options,
		rules:     rules,
	}

	sint64Field := &SInt64Field{}
	sint64Field.ProtoField = &ProtoField[SInt64Field]{
		protoFieldInternal: internal,
		self:               sint64Field,
	}
	sint64Field.NumericField = newNumericField[SInt64Field, int64](internal, sint64Field, false)
	return sint64Field
}

// An instance of a fixed32 protobuf field.
type Fixed32Field struct {
	*NumericField[Fixed32Field, uint32]
}

// The constructor for a fixed32 protobuf field.
func Fixed32(name string) *Fixed32Field {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "fixed32",
		goType:    "uint32",
		options:   options,
		rules:     rules,
	}

	fixed32Field := &Fixed32Field{}
	fixed32Field.ProtoField = &ProtoField[Fixed32Field]{
		protoFieldInternal: internal,
		self:               fixed32Field,
	}
	fixed32Field.NumericField = newNumericField[Fixed32Field, uint32](internal, fixed32Field, false)
	return fixed32Field
}

// An instance of a fixed64 protobuf field.
type Fixed64Field struct {
	*NumericField[Fixed64Field, uint64]
}

// The constructor for a fixed64 protobuf field.
func Fixed64(name string) *Fixed64Field {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "fixed64",
		goType:    "uint64",
		options:   options,
		rules:     rules,
	}

	fixed64Field := &Fixed64Field{}
	fixed64Field.ProtoField = &ProtoField[Fixed64Field]{
		protoFieldInternal: internal,
		self:               fixed64Field,
	}
	fixed64Field.NumericField = newNumericField[Fixed64Field, uint64](internal, fixed64Field, false)
	return fixed64Field
}

// An instance of a sfixed32 protobuf field.
type SFixed32Field struct {
	*NumericField[SFixed32Field, int32]
}

// The constructor for a sfixed32 protobuf field.
func SFixed32(name string) *SFixed32Field {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "sfixed32",
		goType:    "int32",
		options:   options,
		rules:     rules,
	}

	sfixed32Field := &SFixed32Field{}
	sfixed32Field.ProtoField = &ProtoField[SFixed32Field]{
		protoFieldInternal: internal,
		self:               sfixed32Field,
	}
	sfixed32Field.NumericField = newNumericField[SFixed32Field, int32](internal, sfixed32Field, false)
	return sfixed32Field
}

// An instance of a sfixed64 protobuf field.
type SFixed64Field struct {
	*NumericField[SFixed64Field, int64]
}

// The constructor for a sfixed64 protobuf field.
func SFixed64(name string) *SFixed64Field {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "sfixed64",
		goType:    "int64",
		options:   options,
		rules:     rules,
	}

	sfixed64Field := &SFixed64Field{}
	sfixed64Field.ProtoField = &ProtoField[SFixed64Field]{
		protoFieldInternal: internal,
		self:               sfixed64Field,
	}
	sfixed64Field.NumericField = newNumericField[SFixed64Field, int64](internal, sfixed64Field, false)
	return sfixed64Field
}
