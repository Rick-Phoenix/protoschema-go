package schemabuilder

import (
	"fmt"
	"log"
	"os"

	gofirst "github.com/Rick-Phoenix/gofirst/db/queries/gen"
)

type UserWithPosts struct {
	ID        int64          `json:"id"`
	Name      string         `json:"name"`
	CreatedAt string         `dbignore:"true" json:"created_at"`
	Posts     []gofirst.Post `json:"posts"`
}

var UserSchema = ProtoMessageSchema{
	Name: "User",
	Fields: ProtoFieldsMap{
		1: ProtoString("name"),
	},
	ReservedNames:   ReservedNames("name2", "name3"),
	ReservedNumbers: ReservedNumbers(101, 102),
	ReservedRanges:  []Range{{2010, 2029}, {3050, 3055}},
}

var GetUserSchema = ProtoMessageSchema{
	Name: "GetUserRequest",
	Fields: ProtoFieldsMap{
		1: UserSchema.Fields[1],
	},
}

var SubRedditSchema = ProtoMessageSchema{
	Name: "Subreddit",
	Fields: ProtoFieldsMap{
		1: ProtoInt32("id"),
		2: ProtoString("name").MinLen(1).MaxLen(48),
		3: ProtoString("description").MaxLen(255),
		4: ProtoInt32("creator_id"),
		5: RepeatedField("posts", MessageType[gofirst.Post]("Post", WithImportPath("myapp/v1/Post.proto"))),
		6: ProtoTimestamp("created_at"),
	},
}

var PostSchema = ProtoMessageSchema{
	Fields: ProtoFieldsMap{
		1: ProtoString("id"),
		2: ProtoTimestamp("created_at"),
		3: ProtoInt32("creator_id"),
		4: ProtoString("title"),
		5: ProtoString("content"),
		6: ProtoInt32("subreddit_id"),
	},
}

type ServicesMap map[string]ProtoServiceSchema

type ServicesData map[string]ProtoService

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

var ProtoServices = ServicesMap{
	"User": ProtoServiceSchema{
		Handlers: HandlersMap{
			"GetUser": {GetUserSchema, ProtoMessageSchema{
				Name: "GetUserResponse",
				Fields: ProtoFieldsMap{
					1: MessageType[gofirst.User]("User"),
				},
			}},
			"UpdateUser": {ProtoMessageSchema{Name: "UpdateUserResponse", Fields: ProtoFieldsMap{
				1: FieldMask("field_mask"),
			}}, ProtoEmpty()},
		},
	},
	// "Post": ProtoServiceSchema{
	// 	Messages: []ProtoMessageSchema{PostSchema},
	// 	Handlers: HandlersMap{
	// 		"GetPost": {PickFields(&PostSchema, "GetPostRequest", "id"), ProtoMessageSchema{
	// 			Name: "GetPostResponse",
	// 			Fields: ProtoFieldsMap{
	// 				"post": MessageType[gofirst.Post](1, "Post"),
	// 			},
	// 		}},
	// 	},
	// },
	// "Subreddit": ProtoServiceSchema{
	// 	Messages: []ProtoMessageSchema{SubRedditSchema},
	// 	Handlers: HandlersMap{
	// 		"GetSubreddit": {PickFields(&SubRedditSchema, "GetSubredditRequest", "id"), ProtoMessageSchema{
	// 			Name: "GetSubredditResponse",
	// 			Fields: ProtoFieldsMap{
	// 				"subreddit": MessageType[gofirst.Subreddit](1, "Subreddit"),
	// 			},
	// 		}},
	// 	},
	// },
}

func GenerateProtoFiles() {
	Services := BuildFinalServicesMap(ProtoServices)
	templatePath := "templates/service.proto.tmpl"
	outputRoot := "gen/proto"

	options := &Options{TmplPath: templatePath, ProtoRoot: outputRoot}

	for _, v := range Services {
		if err := Generate(v, *options); err != nil {
			log.Fatalf("🔥 Generation failed: %v", err)
		}
	}
}
