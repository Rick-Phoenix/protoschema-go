package protoschema

// An instance of a boolean protobuf field.
type BoolField struct {
	*ProtoField[BoolField]
	*ConstField[BoolField, bool, bool]
	*OptionalField[BoolField]
}

// The constructor for the protobuf boolean field.
func Bool(name string) *BoolField {
	bf := &BoolField{}
	internal := &protoFieldInternal{
		name: name, protoType: "bool", goType: "bool",
	}
	bf.ProtoField = &ProtoField[BoolField]{internal, bf}
	bf.ConstField = &ConstField[BoolField, bool, bool]{constInternal: internal, self: bf}
	bf.OptionalField = &OptionalField[BoolField]{optionalInternal: internal, self: bf}

	return bf
}
