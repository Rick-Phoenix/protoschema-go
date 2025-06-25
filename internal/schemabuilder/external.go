package schemabuilder

import (
	"log"
	"reflect"
	"strings"
)

func createMsgField(name string, s *ProtoMessageSchema) *GenericField {
	rules := make(map[string]any)

	if s == nil {
		log.Fatalf("Could not generate the message type for field %q because the schema given was nil.", name)
	}

	goType := strings.ToLower(s.Name)

	if s.Model != nil {
		goType = reflect.TypeOf(s.Model).Elem().String()
	}

	if s.Name == "" {
		log.Fatalf("Could not generate the message type for field %q because the schema given has no name.", name)
	}

	imports := []string{}

	if s.ImportPath != "" {
		imports = append(imports, s.ImportPath)
	}

	internal := &protoFieldInternal{
		name:        name,
		protoType:   s.Name,
		goType:      goType,
		isNonScalar: true,
		rules:       rules,
		imports:     imports,
	}

	gf := &GenericField{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[GenericField]{
		protoFieldInternal: internal,
		self:               gf,
	}
	return gf
}

func MsgField(name string, s *ProtoMessageSchema) *GenericField {
	return createMsgField(name, s)
}
