package protoschema

import (
	"log"
	"reflect"
)

// A field that has a protobuf message as its type.
func MsgField(name string, s *MessageSchema) *GenericField {
	rules := make(map[string]any)

	if s == nil {
		log.Fatalf("Could not generate the message type for field %q because the schema given was nil.", name)
	}

	if s.Name == "" {
		log.Fatalf("Could not generate the message type for field %q because the schema given has no name.", name)
	}

	var goType string

	if s.Model != nil {
		goType = reflect.TypeOf(s.Model).String()
	} else {
		goType = "any"
	}

	imports := []string{}

	if importPath := s.GetImportPath(); importPath != "" {
		imports = append(imports, importPath)
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
