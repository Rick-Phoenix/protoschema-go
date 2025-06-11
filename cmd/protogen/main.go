package main

import (
	"log"

	"github.com/Rick-Phoenix/gofirst/internal/schemabuilder"
	"github.com/Rick-Phoenix/gofirst/internal/schemabuilder/protogen"
)

// Our NAMED schema type.
type UserSchema struct {
	ID   schemabuilder.ColumnBuilder[int64]
	Name schemabuilder.ColumnBuilder[string]
}

func main() {
	// Initialize our schema.
	schema := schemabuilder.User2

	// Define paths and options.
	templatePath := "templates/service.proto.tmpl"
	outputRoot := "gen/proto"
	version := "v1"

	options := &protogen.Options{TmplPath: templatePath, ProtoRoot: outputRoot, Version: version, ProjectName: "test"}

	// Run the generator!
	if err := protogen.Generate2(schema, *options); err != nil {
		log.Fatalf("🔥 Generation failed: %v", err)
	}
}
