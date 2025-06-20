package schemabuilder

import (
	"fmt"
	"log"
	"os"

	gofirst "github.com/Rick-Phoenix/gofirst/db/queries/gen"
)

var UserSchema = ProtoMessageSchema{
	Name: "User",
	Fields: ProtoFieldsMap{
		"name":  RepeatedField(1, ProtoString(1).MinLen(2)).Unique().Required().Deprecated().MinItems(10),
		"name2": ProtoString(2).Required().MinLen(2),
		"createdAt": ProtoTimestamp(3).Required().CelOptions(CelFieldOpts{
			Id:         "test",
			Message:    "this is a test",
			Expression: "this = test",
		}),
		"post":    MessageType[gofirst.Post](4, "Post", WithImportPath("myapp/v1/Post.proto")),
		"maptype": ProtoMap(209, ProtoInt32(0).Lt(10), ProtoString(0).Example("aa").Const("aaa")).Required(),
	},
	Oneofs: ProtoOneofsMap{
		"myoneof": ProtoOneOf(OneofChoicesMap{
			"choice1": ProtoString(5),
		})},
	Options: []ProtoOption{DisableValidator, ProtoCustomOneOf(false, "aa", "bb")},
}

var PostSchema = ProtoMessageSchema{
	Fields: ProtoFieldsMap{
		"id": ProtoString(1),
		"createdAt": ProtoTimestamp(2).CelOptions(CelFieldOpts{
			Id:         "test",
			Message:    "this is a test",
			Expression: "this = test",
		}).LtNow(),
	},
}

type ServicesMap map[string]ProtoServiceSchema

type ServicesData map[string]ProtoService

// Make something that reflects the db field names and types and checks if the messages are correct
func BuildFinalServicesMap(m ServicesMap) ServicesData {
	out := make(ServicesData)
	serviceErrors := []error{}

	for resource, serviceSchema := range m {
		serviceData, err := NewProtoService(resource, serviceSchema, "myapp/v1")

		if err != nil {
			serviceErrors = append(serviceErrors, fmt.Errorf("Errors for the schema %s:\n%s", resource, IndentString(err.Error())))
		}
		out[resource] = serviceData
	}

	if len(serviceErrors) > 0 {
		fmt.Printf("The following errors occurred:\n\n")
		for _, err := range serviceErrors {
			fmt.Println(err)
		}

		os.Exit(1)
	}

	return out
}

var MyOptions = []CustomOption{{
	Name: "testopt", Type: "string", FieldNr: 1, Optional: true,
}}

// Maybe separate db types from other messages (which can add or subtract from db types)
// Better to have message names in their own schemas. Remove the map here and use generator functions instead
// like NewService(name, schema)
// Then aggregate them in a slice (or define them directly in it) to generate them

var TablesData = ServicesMap{
	"User": ProtoServiceSchema{
		Handlers: HandlersMap{
			"GetUser":    {ProtoEmpty(), UserSchema},
			"UpdateUser": {MessageRef("UpdateUserResponse"), ProtoEmpty()},
		},
	},
	"Post": ProtoServiceSchema{
		Messages: []ProtoMessageSchema{PostSchema},
		Handlers: HandlersMap{
			"GetPost": {PostSchema, PostSchema},
		},
	},
}

func GenerateProtoFiles() {
	var Services = BuildFinalServicesMap(TablesData)
	templatePath := "templates/service.proto.tmpl"
	outputRoot := "gen/proto"

	options := &Options{TmplPath: templatePath, ProtoRoot: outputRoot}

	for _, v := range Services {
		if err := Generate(v, *options); err != nil {
			log.Fatalf("🔥 Generation failed: %v", err)
		}
	}
}
