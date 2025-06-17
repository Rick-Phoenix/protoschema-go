package schemabuilder

import "reflect"

func MessageType[ValueT any](fieldNr int, name string, opts ...FieldPathGetter) *GenericField[ValueT] {
	imports := make(Set)
	rules := make(map[string]any)

	internal := &protoFieldInternal{
		fieldNr:     fieldNr,
		protoType:   name,
		goType:      reflect.TypeOf((*ValueT)(nil)).Elem().String(),
		isNonScalar: true,
		imports:     imports,
		rules:       rules,
	}

	for _, opt := range opts {
		opt(internal)
	}

	gf := &GenericField[ValueT]{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[GenericField[ValueT], ValueT]{
		protoFieldInternal: internal,
		self:               gf,
	}
	return gf
}

type FieldPathGetter func(*protoFieldInternal)

func WithImportPath(path string) FieldPathGetter {
	return func(pfi *protoFieldInternal) {
		pfi.imports[path] = present
	}
}
