package main

import (
	"log"

	sb "github.com/Rick-Phoenix/gofirst/internal/schemabuilder"
	"github.com/Rick-Phoenix/gofirst/internal/schemabuilder/protogen"
)

var UserSchema = &sb.TableBuilder{
	Name: "User",
	Columns: sb.ColumnsMap{
		"Name":      sb.StringCol().Required().MinLen(3).Requests("create").Responses("get", "create").Nullable(),
		"Age":       sb.Int64Col().Responses("get").Nullable(),
		"Blob":      sb.BytesCol().Requests("get"),
		"CreatedAt": sb.TimestampCol().Responses("get"),
	},
}

func main() {

	// Define paths and options.
	templatePath := "templates/service.proto.tmpl"
	outputRoot := "gen/proto"
	version := "v1"

	options := &protogen.Options{TmplPath: templatePath, ProtoRoot: outputRoot, Version: version, ProjectName: "test"}

	// Run the generator!
	if err := protogen.Generate(UserSchema, *options); err != nil {
		log.Fatalf("🔥 Generation failed: %v", err)
	}
}
