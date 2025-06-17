package schemabuilder

import (
	"fmt"
	"log"
	"os"
)

type FieldData map[string]*ServiceData

// Oneof should not accept optional or required fields
// Applying deprecated to services and messages
// Repeated = true for repeated options
// Or separate array
var UserSchema = ProtoMessageSchema{
	Fields: ProtoFieldsMap{
		"name":  RepeatedField(ProtoString(1).MinLen(2)).Unique(),
		"name2": ProtoString(2).Required().MinLen(2),
		"createdAt": ProtoTimestamp(3).Required().CelField(CelFieldOpts{
			Id:         "test",
			Message:    "this is a test",
			Expression: "this = test",
		}),
		"post": MessageType(4, "Post", "myapp/v1/Post.proto"),
	},
	OneOfs: []ProtoOneOfSchema{{
		Name: "myoneof", Options: []ProtoOption{{Name: "myopt", Value: "true"}, OneOfRequired}, Choices: ProtoOneOfsMap{
			"choice1": ProtoString(5),
			"choice2": ProtoInt(6),
		},
	}},
}

var PostSchema = ProtoMessageSchema{
	Fields: ProtoFieldsMap{
		"id": ProtoString(1),
		"createdAt": ProtoTimestamp(2).CelField(CelFieldOpts{
			Id:         "test",
			Message:    "this is a test",
			Expression: "this = test",
		}).Optional(),
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
					"user": InternalType(1, "User"),
					"createdAt": ProtoTimestamp(2).Required().CelField(CelFieldOpts{
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
					"user": MessageType(1, "User", "myapp/v1/User.proto"),
					"post": InternalType(2, "Post"),
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
