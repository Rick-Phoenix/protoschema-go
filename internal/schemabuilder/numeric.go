package schemabuilder

import "fmt"

type NumericField[BuilderT any, ValueT any] struct {
	*protoFieldInternal
	self *BuilderT

	hasLtOrLte bool
	hasGtOrGte bool

	isFloatType bool
}

func newNumericField[BuilderT any, ValueT any](pfi *protoFieldInternal, self *BuilderT, isFloat bool) *NumericField[BuilderT, ValueT] {
	return &NumericField[BuilderT, ValueT]{
		protoFieldInternal: pfi,
		self:               self,
		isFloatType:        isFloat,
	}
}

func (nf *NumericField[BuilderT, ValueT]) Lt(val ValueT) *BuilderT {
	if nf.hasLtOrLte {
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

func (nf *NumericField[BuilderT, ValueT]) In(vals ...ValueT) *BuilderT {
	nf.rules["in"] = vals
	return nf.self
}

func (nf *NumericField[BuilderT, ValueT]) NotIn(vals ...ValueT) *BuilderT {
	nf.rules["not_in"] = vals
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

func ProtoInt(fieldNumber int) *IntField {
	imports := make(Set)
	options := make(map[string]string)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		fieldNr:   fieldNumber,
		protoType: "int32",
		goType:    "int32",
		imports:   imports,
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

func ProtoFloat(fieldNumber int) *FloatField {
	imports := make(Set)
	options := make(map[string]string)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		fieldNr:   fieldNumber,
		protoType: "float",
		goType:    "float32",
		imports:   imports,
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
