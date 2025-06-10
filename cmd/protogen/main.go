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
	schema := &UserSchema{
		ID:   schemabuilder.Int64Col(),
		Name: schemabuilder.StringCol().Required().MinLen(3),
	}

	// Define paths and options.
	templatePath := "templates/service.proto.tmpl"
	outputRoot := "gen/proto"
	version := "v1"

	// Run the generator!
	if err := protogen.Generate(schema, templatePath, outputRoot, version); err != nil {
		log.Fatalf("🔥 Generation failed: %v", err)
	}
}
