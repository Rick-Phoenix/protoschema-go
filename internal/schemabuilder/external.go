package schemabuilder

import "reflect"

func MessageType[ValueT any](name string, protoType string, importPath string) *GenericField[ValueT] {
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		name:        name,
		protoType:   protoType,
		goType:      reflect.TypeOf((*ValueT)(nil)).Elem().String(),
		isNonScalar: true,
		rules:       rules,
	}

	if importPath != "" {
		internal.imports = append(internal.imports, importPath)
	}

	gf := &GenericField[ValueT]{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[GenericField[ValueT], ValueT]{
		protoFieldInternal: internal,
		self:               gf,
	}
	return gf
}

func MsgField[ValueT any](name, protoType string) *GenericField[ValueT] {
	return MessageType[ValueT](name, protoType, "")
}

func ImportedMsgField[ValueT any](name, protoType, importPath string) *GenericField[ValueT] {
	return MessageType[ValueT](name, protoType, importPath)
}
