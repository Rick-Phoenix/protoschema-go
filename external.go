package schemabuilder

import (
	"log"
	"reflect"
)

func MsgField(name string, s *MessageSchema) *GenericField {
	rules := make(map[string]any)

	if s == nil {
		log.Fatalf("Could not generate the message type for field %q because the schema given was nil.", name)
	}

	var goType string

	if s.Model == nil {
		goType = "any"
	} else {
		goType = reflect.TypeOf(s.Model).String()
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
		messageRef:  s,
	}

	gf := &GenericField{}
	gf.ProtoField = &ProtoField[GenericField]{
		protoFieldInternal: internal,
		self:               gf,
	}
	return gf
}
