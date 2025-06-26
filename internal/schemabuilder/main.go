package schemabuilder

import (
	"fmt"
	"log"
	"os"
	"time"

	gofirst "github.com/Rick-Phoenix/gofirst/db/queries/gen"
)

type UserWithPosts struct {
	ID        int64          `json:"id"`
	Name      string         `json:"name"`
	CreatedAt time.Time      `dbignore:"true" json:"created_at"`
	Posts     []gofirst.Post `json:"posts"`
}

var UserSchema = ProtoMessageSchema{
	Name: "User",
	Fields: ProtoFieldsMap{
		1: ProtoString("name"),
		2: ProtoInt64("id"),
		3: ProtoTimestamp("created_at"),
		5: RepeatedField("posts", MsgField("post", &PostSchema)).CelOption("myexpr", "x must not be y", "x != y"),
	},
	Oneofs: []ProtoOneOfBuilder{ProtoOneOf("myoneof", OneofChoicesMap{
		9:  ProtoString("example"),
		10: ProtoInt32("another"),
	}, []ProtoOption{{"myopt1", "myval1"}, {"myopt", "myval"}}...)},
	Enums: []ProtoEnumGroup{
		ProtoEnum("myenum", ProtoEnumMap{
			0: "VAL_1",
			1: "VAL_2",
		}).Opts([]ProtoOption{{"myopt1", "myval1"}, {"myopt", "myval"}}...),
	},
	Model:      &UserWithPosts{},
	ImportPath: "myapp/v1/user.proto",
	Options:    []ProtoOption{MessageCelOption("myexpr", "x must not be y", "x != y")},
}

var GetUserSchema = ProtoMessageSchema{
	Name: "GetUserRequest",
	Fields: ProtoFieldsMap{
		1: UserSchema.GetField("name"),
	},
}

var GetPostSchema = ProtoMessageSchema{
	Name: "GetPostRequest",
	Fields: ProtoFieldsMap{
		1: PostSchema.GetField("id"),
	},
}

var SubRedditSchema = ProtoMessageSchema{
	Name: "Subreddit",
	Fields: ProtoFieldsMap{
		1: ProtoInt32("id").Optional(),
		2: ProtoString("name").MinLen(1).MaxLen(48),
		3: ProtoString("description").MaxLen(255),
		4: ProtoInt32("creator_id"),
		5: RepeatedField("posts", MsgField("post", &PostSchema)),
		6: ProtoTimestamp("created_at"),
	},
}

var PostSchema = ProtoMessageSchema{
	Name: "Post",
	Fields: ProtoFieldsMap{
		1: ProtoInt64("id").Optional(),
		2: ProtoTimestamp("created_at"),
		3: ProtoInt64("author_id"),
		4: ProtoString("title").MinLen(5).MaxLen(64).Required(),
		5: ProtoString("content").Optional(),
		6: ProtoInt64("subreddit_id"),
	},
	ImportPath:  "myapp/v1/post.proto",
	ModelIgnore: []string{"content"},
	Model:       &gofirst.Post{},
}

type ServicesMap map[string]ProtoServiceSchema

type ServicesData map[string]ProtoService

var UserService = ProtoServiceSchema{
	Messages: []ProtoMessageSchema{UserSchema},
	Handlers: HandlersMap{
		"GetUser": {GetUserSchema, ProtoMessageSchema{
			Name: "GetUserResponse",
			Fields: ProtoFieldsMap{
				1: MsgField("user", &UserSchema),
			},
		}},
		"UpdateUser": {ProtoMessageSchema{Name: "UpdateUserRequest", Fields: ProtoFieldsMap{
			1: FieldMask("field_mask"),
			2: MsgField("user", &UserSchema),
		}}, ProtoEmpty()},
	},
	FileOptions:    []ProtoOption{{"myopt1", "myval1"}, {"myopt", "myval"}},
	ServiceOptions: []ProtoOption{{"myopt1", "myval1"}, {"myopt", "myval"}},
}

var ProtoServices = ServicesMap{
	"User": UserService,
	"Post": ProtoServiceSchema{
		Messages: []ProtoMessageSchema{PostSchema},
		Handlers: HandlersMap{
			"GetPost": {GetPostSchema, ProtoMessageSchema{
				Name: "GetPostResponse",
				Fields: ProtoFieldsMap{
					1: MsgField("post", &PostSchema),
				},
			}},
			"UpdatePost": {ProtoMessageSchema{Name: "UpdatePostRequest", Fields: ProtoFieldsMap{
				1: MsgField("post", &PostSchema),
				2: FieldMask("field_mask"),
			}}, ProtoEmpty()},
		},
	},
}

func GenerateProtoFiles() {
	Services := BuildServicesMap(ProtoServices)
	outputRoot := "gen/proto"

	options := Options{ProtoRoot: outputRoot}

	for _, v := range Services {
		if err := GenerateProtoFile(v, options); err != nil {
			log.Fatal(err)
		}
	}
}

func BuildServicesMap(m ServicesMap) ServicesData {
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
