package main

import (
	"log"

	"github.com/Rick-Phoenix/gofirst/internal/schemabuilder"
	"github.com/Rick-Phoenix/gofirst/internal/schemabuilder/protogen"
)

func main() {
	// Initialize our schema.
	schema := schemabuilder.UserSchema

	// Define paths and options.
	templatePath := "templates/service.proto.tmpl"
	outputRoot := "gen/proto"
	version := "v1"

	options := &protogen.Options{TmplPath: templatePath, ProtoRoot: outputRoot, Version: version, ProjectName: "test"}

	// Run the generator!
	if err := protogen.Generate(schema, *options); err != nil {
		log.Fatalf("🔥 Generation failed: %v", err)
	}
}
