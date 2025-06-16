package main

import (
	"log"

	sb "github.com/Rick-Phoenix/gofirst/internal/schemabuilder"
)

func main() {

	// Define paths and options.
	templatePath := "templates/service.proto.tmpl"
	outputRoot := "gen/proto"

	options := &sb.Options{TmplPath: templatePath, ProtoRoot: outputRoot}

	// Run the generator!
	if err := sb.Generate(sb.UserService, *options); err != nil {
		log.Fatalf("🔥 Generation failed: %v", err)
	}
}
