package schemabuilder

import (
	"log"
	"reflect"
)

func createMsgField(name string, s *ProtoMessageSchema, withImport bool) *GenericField[any] {
	rules := make(map[string]any)

	if s == nil {
		log.Fatalf("Could not generate the message type for field %q because the schema given was nil.", name)
	}

	if s.Model == nil {
		log.Fatalf("Message field %q has no model to refer to.", name)
	}

	if s.Name == "" {
		log.Fatalf("Could not generate the message type for field %q because the schema given was nil.", name)
	}

	imports := []string{}

	if withImport {
		if s.ImportPath == "" {
			log.Fatalf("Message field %q is missing an import path.", name)
		}
		imports = append(imports, s.ImportPath)
	}

	internal := &protoFieldInternal{
		name:        name,
		protoType:   s.Name,
		goType:      reflect.TypeOf(s.Model).Elem().String(),
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

func ImportedMsgField(name string, s *ProtoMessageSchema) *GenericField[any] {
	return createMsgField(name, s, true)
}

func MsgField(name string, s *ProtoMessageSchema) *GenericField[any] {
	return createMsgField(name, s, false)
}
