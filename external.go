package schemabuilder

import (
	"fmt"
	"log"
	"reflect"
)

func MsgField(name string, s *MessageSchema) *ProtoGenericField {
	rules := make(map[string]any)

	if s == nil {
		log.Fatalf("Could not generate the message type for field %q because the schema given was nil.", name)
	}

	var goType string

	if s.Model == nil {
		fmt.Printf("Message schema %q referenced in field %q has no model to extract the go type from. Using 'any' as a fallback...\n", s.Name, name)
		goType = "any"
	} else {
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

	gf := &ProtoGenericField{}
	gf.ProtoFieldExternal = &ProtoFieldExternal[ProtoGenericField]{
		protoFieldInternal: internal,
		self:               gf,
	}
	return gf
}
