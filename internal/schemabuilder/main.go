package schemabuilder

import "log"

type FieldData map[string]*ServiceData

// Reusable handlers that give the shared functions like len
// Oneof
// Make specific builders that have their own methods (embed the generic interface)
// Add ignore options
var UserSchema = ProtoMessageSchema{
	Fields: ProtoFieldsMap{
		"name": ProtoString(1).MinLen(2),
		"createdAt": ProtoTimestamp(2).Required().CelField(CelFieldOpts{
			Id:         "test",
			Message:    "this is a test",
			Expression: "this = test",
		}).Optional(),
		"post": ImportedType(3, "Post", "myapp/v1/Post.proto").Repeated(),
	},
}

var PostSchema = ProtoMessageSchema{
	Fields: ProtoFieldsMap{
		"id": ProtoString(1),
		"createdAt": ProtoTimestamp(2).Required().CelField(CelFieldOpts{
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

	for resource, serviceSchema := range m {
		out[resource] = NewProtoService(resource, serviceSchema, "myapp/v1")
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
			Service: MyOptions,
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
					"user": ImportedType(1, "User", "myapp/v1/User.proto"),
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
