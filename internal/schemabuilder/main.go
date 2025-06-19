package schemabuilder

import (
	"fmt"
	"log"
	"os"

	gofirst "github.com/Rick-Phoenix/gofirst/db/queries/gen"
)

type FieldData map[string]*ServiceData

// Read again how oneof goes with optional and required
// Applying deprecated to services and messages
// (buf.validate.message).oneof
// Change service response/request definition to allow Empty
var UserSchema = ProtoMessageSchema{
	Fields: ProtoFieldsMap{
		"name":  RepeatedField(1, ProtoString(1).MinLen(2)).Unique(),
		"name2": ProtoString(2).Required().MinLen(2),
		"createdAt": ProtoTimestamp(3).Required().CelOption(CelFieldOpts{
			Id:         "test",
			Message:    "this is a test",
			Expression: "this = test",
		}),
		"post":    MessageType[gofirst.Post](4, "Post", WithImportPath("myapp/v1/Post.proto")),
		"maptype": ProtoMap(5, ProtoInt32(0).Lt(10), ProtoString(0).Example("aa").Const("aaa")),
	},
	OneOfs: []ProtoOneOfBuilder{
		ProtoOneOf("myoneof", ProtoOneOfsMap{
			"choice1": ProtoString(5),
			// "choice2": ProtoInt(6),
		})},
}

var PostSchema = ProtoMessageSchema{
	Fields: ProtoFieldsMap{
		"id": ProtoString(1),
		"createdAt": ProtoTimestamp(2).CelOption(CelFieldOpts{
			Id:         "test",
			Message:    "this is a test",
			Expression: "this = test",
		}),
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

var TablesData = ServicesMap{
	"User": ProtoServiceSchema{
		OptionExtensions: OptionExtensions{
			Field:   MyOptions,
			Message: MyOptions,
			File:    MyOptions,
			OneOf:   MyOptions,
		},
		FileOptions:    []ProtoOption{{Name: "myoption", Value: "true"}},
		ServiceOptions: []ProtoOption{{Name: "myoption", Value: "true"}},
		Resource:       UserSchema,
		Get: &ServiceData{
			Request: UserSchema,
			Response: ProtoMessageSchema{
				Reserved: []int{100, 101, 102},
				Fields: ProtoFieldsMap{
					"user": MessageType[gofirst.User](1, "User"),
					"createdAt": ProtoTimestamp(2).Required().CelOption(CelFieldOpts{
						Id:         "test",
						Message:    "this is a test",
						Expression: "this = test",
					}),
				},
			},
		},
	},
	"Post": ProtoServiceSchema{
		Resource: PostSchema,
		Get: &ServiceData{
			Request: PostSchema,
			Response: ProtoMessageSchema{
				Reserved: []int{100, 101, 102},
				Fields: ProtoFieldsMap{
					"user": MessageType[gofirst.User](1, "User", WithImportPath("myapp/v1/Post.proto")),
					"post": MessageType[gofirst.Post](2, "Post"),
				},
			},
		},
	},
}

func GenerateProtoFiles() {
	var Services = BuildFinalServicesMap(TablesData)
	// Define paths and options.
	templatePath := "templates/service.proto.tmpl"
	outputRoot := "gen/proto"

	options := &Options{TmplPath: templatePath, ProtoRoot: outputRoot}

	for _, v := range Services {
		if err := Generate(v, *options); err != nil {
			log.Fatalf("🔥 Generation failed: %v", err)
		}
	}
}
