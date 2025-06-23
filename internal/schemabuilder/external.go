package schemabuilder

import (
	"log"
	"reflect"
)

func MsgField(name string, s *ProtoMessageSchema) *GenericField[any] {
	rules := make(map[string]any)

	if s == nil {
		log.Fatalf("Could not generate the message type for field %q because the schema given was nil.", name)
	}

	if s.DbModel == nil {
		log.Fatalf("Message field %q has no model to refer to.", name)
	}

	if s.Name == "" {
		log.Fatalf("Could not generate the message type for field %q because the schema given was nil.", name)
	}

	imports := []string{}

	if s.ImportPath != "" {
		imports = append(imports, s.ImportPath)
	}

	internal := &protoFieldInternal{
		name:        name,
		protoType:   s.Name,
		goType:      reflect.TypeOf(s.DbModel).Elem().String(),
		isNonScalar: true,
		rules:       rules,
		imports:     imports,
	}

	gf := &GenericField[any]{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[GenericField[any], any]{
		protoFieldInternal: internal,
		self:               gf,
	}
	return gf
}
