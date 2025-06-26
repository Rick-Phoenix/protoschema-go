package schemabuilder

import (
	"errors"
	"fmt"

	"golang.org/x/exp/constraints"
)

type ProtoNumericField[BuilderT any, ValueT constraints.Ordered] struct {
	*ProtoFieldExternal[BuilderT]
	*ProtoConstField[BuilderT, ValueT, ValueT]
	*ProtoOptionalField[BuilderT]
	self *BuilderT

	hasLtOrLte bool
	hasGtOrGte bool

	lt  *ValueT
	lte *ValueT
	gt  *ValueT
	gte *ValueT

	isFloatType bool
}

func newNumericField[BuilderT any, ValueT constraints.Ordered](pfi *protoFieldInternal, self *BuilderT, isFloat bool) *ProtoNumericField[BuilderT, ValueT] {
	return &ProtoNumericField[BuilderT, ValueT]{
		ProtoFieldExternal: &ProtoFieldExternal[BuilderT]{
			protoFieldInternal: pfi,
			self:               self,
		},
		isFloatType: isFloat,
		ProtoConstField: &ProtoConstField[BuilderT, ValueT, ValueT]{
			constInternal: pfi,
			self:          self,
		},
		ProtoOptionalField: &ProtoOptionalField[BuilderT]{
			optionalInternal: pfi,
			self:             self,
		},
	}
}

func (nf *ProtoNumericField[BuilderT, ValueT]) Lt(val ValueT) *BuilderT {
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
	return nf.ProtoFieldExternal.self
}

func (nf *ProtoNumericField[BuilderT, ValueT]) Lte(val ValueT) *BuilderT {
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
	return nf.ProtoFieldExternal.self
}

func (nf *ProtoNumericField[BuilderT, ValueT]) Gt(val ValueT) *BuilderT {
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
	return nf.ProtoFieldExternal.self
}

func (nf *ProtoNumericField[BuilderT, ValueT]) Gte(val ValueT) *BuilderT {
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
	return nf.ProtoFieldExternal.self
}

func (nf *ProtoNumericField[BuilderT, ValueT]) Finite() *BuilderT {
	if !nf.isFloatType {
		nf.errors = errors.Join(nf.errors, fmt.Errorf("The 'finite' rule is only applicable to float and double types."))
	}
	nf.rules["finite"] = true
	return nf.ProtoFieldExternal.self
}

type ProtoInt32Field struct {
	*ProtoNumericField[ProtoInt32Field, int32]
}

func Int32(name string) *ProtoInt32Field {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "int32",
		goType:    "int32",
		options:   options,
		rules:     rules,
	}

	integerField := &ProtoInt32Field{}

	integerField.ProtoNumericField = newNumericField[ProtoInt32Field, int32](internal, integerField, false)
	return integerField
}

type ProtoFloatField struct {
	*ProtoFieldExternal[ProtoFloatField]
	*ProtoNumericField[ProtoFloatField, float32]
}

func Float(name string) *ProtoFloatField {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "float",
		goType:    "float32",
		options:   options,
		rules:     rules,
	}

	floatField := &ProtoFloatField{}
	floatField.ProtoFieldExternal = &ProtoFieldExternal[ProtoFloatField]{
		protoFieldInternal: internal,
		self:               floatField,
	}
	floatField.ProtoNumericField = newNumericField[ProtoFloatField, float32](internal, floatField, true)
	return floatField
}

type ProtoDoubleField struct {
	*ProtoFieldExternal[ProtoDoubleField]
	*ProtoNumericField[ProtoDoubleField, float64]
}

func Double(name string) *ProtoDoubleField {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "double",
		goType:    "float64",
		options:   options,
		rules:     rules,
	}

	floatField := &ProtoDoubleField{}
	floatField.ProtoFieldExternal = &ProtoFieldExternal[ProtoDoubleField]{
		protoFieldInternal: internal,
		self:               floatField,
	}
	floatField.ProtoNumericField = newNumericField[ProtoDoubleField, float64](internal, floatField, true)
	return floatField
}

type ProtoInt64Field struct {
	*ProtoFieldExternal[ProtoInt64Field]
	*ProtoNumericField[ProtoInt64Field, int64]
}

func Int64(name string) *ProtoInt64Field {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "int64",
		goType:    "int64",
		options:   options,
		rules:     rules,
	}

	int64Field := &ProtoInt64Field{}
	int64Field.ProtoFieldExternal = &ProtoFieldExternal[ProtoInt64Field]{
		protoFieldInternal: internal,
		self:               int64Field,
	}
	int64Field.ProtoNumericField = newNumericField[ProtoInt64Field, int64](internal, int64Field, false)
	return int64Field
}

type ProtoUInt32Field struct {
	*ProtoFieldExternal[ProtoUInt32Field]
	*ProtoNumericField[ProtoUInt32Field, uint32]
}

func UInt32(name string) *ProtoUInt32Field {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "uint32",
		goType:    "uint32",
		options:   options,
		rules:     rules,
	}

	uint32Field := &ProtoUInt32Field{}
	uint32Field.ProtoFieldExternal = &ProtoFieldExternal[ProtoUInt32Field]{
		protoFieldInternal: internal,
		self:               uint32Field,
	}
	uint32Field.ProtoNumericField = newNumericField[ProtoUInt32Field, uint32](internal, uint32Field, false)
	return uint32Field
}

type ProtoUInt64Field struct {
	*ProtoFieldExternal[ProtoUInt64Field]
	*ProtoNumericField[ProtoUInt64Field, uint64]
}

func UInt64(name string) *ProtoUInt64Field {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "uint64",
		goType:    "uint64",
		options:   options,
		rules:     rules,
	}

	uint64Field := &ProtoUInt64Field{}
	uint64Field.ProtoFieldExternal = &ProtoFieldExternal[ProtoUInt64Field]{
		protoFieldInternal: internal,
		self:               uint64Field,
	}
	uint64Field.ProtoNumericField = newNumericField[ProtoUInt64Field, uint64](internal, uint64Field, false)
	return uint64Field
}

type ProtoSInt32Field struct {
	*ProtoFieldExternal[ProtoSInt32Field]
	*ProtoNumericField[ProtoSInt32Field, int32]
}

func SInt32(name string) *ProtoSInt32Field {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "sint32",
		goType:    "int32",
		options:   options,
		rules:     rules,
	}

	sint32Field := &ProtoSInt32Field{}
	sint32Field.ProtoFieldExternal = &ProtoFieldExternal[ProtoSInt32Field]{
		protoFieldInternal: internal,
		self:               sint32Field,
	}
	sint32Field.ProtoNumericField = newNumericField[ProtoSInt32Field, int32](internal, sint32Field, false)
	return sint32Field
}

type ProtoSInt64Field struct {
	*ProtoFieldExternal[ProtoSInt64Field]
	*ProtoNumericField[ProtoSInt64Field, int64]
}

func SInt64(name string) *ProtoSInt64Field {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "sint64",
		goType:    "int64",
		options:   options,
		rules:     rules,
	}

	sint64Field := &ProtoSInt64Field{}
	sint64Field.ProtoFieldExternal = &ProtoFieldExternal[ProtoSInt64Field]{
		protoFieldInternal: internal,
		self:               sint64Field,
	}
	sint64Field.ProtoNumericField = newNumericField[ProtoSInt64Field, int64](internal, sint64Field, false)
	return sint64Field
}

type ProtoFixed32Field struct {
	*ProtoFieldExternal[ProtoFixed32Field]
	*ProtoNumericField[ProtoFixed32Field, uint32]
}

func Fixed32(name string) *ProtoFixed32Field {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "fixed32",
		goType:    "uint32",
		options:   options,
		rules:     rules,
	}

	fixed32Field := &ProtoFixed32Field{}
	fixed32Field.ProtoFieldExternal = &ProtoFieldExternal[ProtoFixed32Field]{
		protoFieldInternal: internal,
		self:               fixed32Field,
	}
	fixed32Field.ProtoNumericField = newNumericField[ProtoFixed32Field, uint32](internal, fixed32Field, false)
	return fixed32Field
}

type ProtoFixed64Field struct {
	*ProtoFieldExternal[ProtoFixed64Field]
	*ProtoNumericField[ProtoFixed64Field, uint64]
}

func Fixed64(name string) *ProtoFixed64Field {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "fixed64",
		goType:    "uint64",
		options:   options,
		rules:     rules,
	}

	fixed64Field := &ProtoFixed64Field{}
	fixed64Field.ProtoFieldExternal = &ProtoFieldExternal[ProtoFixed64Field]{
		protoFieldInternal: internal,
		self:               fixed64Field,
	}
	fixed64Field.ProtoNumericField = newNumericField[ProtoFixed64Field, uint64](internal, fixed64Field, false)
	return fixed64Field
}

type ProtoSFixed32Field struct {
	*ProtoFieldExternal[ProtoSFixed32Field]
	*ProtoNumericField[ProtoSFixed32Field, int32]
}

func SFixed32(name string) *ProtoSFixed32Field {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "sfixed32",
		goType:    "int32",
		options:   options,
		rules:     rules,
	}

	sfixed32Field := &ProtoSFixed32Field{}
	sfixed32Field.ProtoFieldExternal = &ProtoFieldExternal[ProtoSFixed32Field]{
		protoFieldInternal: internal,
		self:               sfixed32Field,
	}
	sfixed32Field.ProtoNumericField = newNumericField[ProtoSFixed32Field, int32](internal, sfixed32Field, false)
	return sfixed32Field
}

type ProtoSFixed64Field struct {
	*ProtoFieldExternal[ProtoSFixed64Field]
	*ProtoNumericField[ProtoSFixed64Field, int64]
}

func SFixed64(name string) *ProtoSFixed64Field {
	options := make(map[string]any)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:      name,
		protoType: "sfixed64",
		goType:    "int64",
		options:   options,
		rules:     rules,
	}

	sfixed64Field := &ProtoSFixed64Field{}
	sfixed64Field.ProtoFieldExternal = &ProtoFieldExternal[ProtoSFixed64Field]{
		protoFieldInternal: internal,
		self:               sfixed64Field,
	}
	sfixed64Field.ProtoNumericField = newNumericField[ProtoSFixed64Field, int64](internal, sfixed64Field, false)
	return sfixed64Field
}
