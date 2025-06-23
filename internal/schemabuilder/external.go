package schemabuilder

import (
	"log"
	"reflect"
)

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

func InternalMsgField(name string, s *ProtoMessageSchema) *GenericField[any] {
	rules := make(map[string]any)

	if s == nil {
		log.Fatalf("Could not generate the message type for field %q because the schema given was nil.", name)
	}

	internal := &protoFieldInternal{
		name:        name,
		protoType:   s.Name,
		goType:      reflect.TypeOf(s.DbModel).Elem().String(),
		isNonScalar: true,
		rules:       rules,
	}

	gf := &GenericField[any]{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[GenericField[any], any]{
		protoFieldInternal: internal,
		self:               gf,
	}
	return gf
}

func ExternalMsgField(name string, s *ProtoMessageSchema) *GenericField[any] {
	rules := make(map[string]any)

	if s == nil {
		log.Fatalf("Could not generate the message type for field %q because the schema given was nil.", name)
	}

	if s.DbModel == nil {
		log.Fatalf("Message field %q has no model to refer to.", name)
	}

	if s.ImportPath == "" {
		log.Fatalf("Message field %q is missing an import path.", name)
	}

	internal := &protoFieldInternal{
		name:        name,
		protoType:   s.Name,
		goType:      reflect.TypeOf(s.DbModel).Elem().String(),
		isNonScalar: true,
		rules:       rules,
		imports:     []string{s.ImportPath},
	}

	gf := &GenericField[any]{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[GenericField[any], any]{
		protoFieldInternal: internal,
		self:               gf,
	}
	return gf
}
