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
		1: ProtoString("name").MinLen(2).MaxLen(48),
		2: ProtoInt64("id"),
		3: ProtoTimestamp("created_at"),
		5: RepeatedField("posts", ImportedMsgField("post", &PostSchema)),
	},
	ReservedNames:   ReservedNames("name2", "name3"),
	ReservedNumbers: ReservedNumbers(101, 102),
	ReservedRanges:  []Range{{2010, 2029}, {3050, 3055}},
	Model:           &gofirst.User{},
	ModelIgnore:     []string{"posts"},
	ImportPath:      "myapp/v1/user.proto",
}

var GetUserSchema = ProtoMessageSchema{
	Name: "GetUserRequest",
	Fields: ProtoFieldsMap{
		1: UserSchema.GetField("name"),
	},
}

var impfield = ProtoString("field")

var GetPostSchema = ProtoMessageSchema{
	Name: "GetPostRequest",
	Fields: ProtoFieldsMap{
		1: PostSchema.GetField("id"),
	},
}

var SubRedditSchema = ProtoMessageSchema{
	Name: "Subreddit",
	Fields: ProtoFieldsMap{
		1: ProtoInt32("id"),
		2: ProtoString("name").MinLen(1).MaxLen(48),
		3: ProtoString("description").MaxLen(255),
		4: ProtoInt32("creator_id"),
		5: RepeatedField("posts", ImportedMsgField("post", &PostSchema)),
		6: ProtoTimestamp("created_at"),
	},
}

var PostSchema = ProtoMessageSchema{
	Name: "Post",
	Fields: ProtoFieldsMap{
		1: ProtoInt64("id"),
		2: ProtoTimestamp("created_at"),
		3: ProtoInt64("author_id"),
		4: ProtoString("title").MinLen(5).MaxLen(64).Required(),
		5: ProtoString("content").Optional(),
		6: ProtoInt64("subreddit_id"),
	},
	ImportPath: "myapp/v1/post.proto",
	Model:      &gofirst.Post{},
}

type ServicesMap map[string]ProtoServiceSchema

type ServicesData map[string]ProtoService

var MyOptions = []CustomOption{{
	Name: "testopt", Type: "string", FieldNr: 1, Optional: true,
}}

var ProtoServices = ServicesMap{
	"User": ProtoServiceSchema{
		Messages: []ProtoMessageSchema{UserSchema},
		Handlers: HandlersMap{
			"GetUser": {GetUserSchema, ProtoMessageSchema{
				Name: "GetUserResponse",
				Fields: ProtoFieldsMap{
					1: MsgField("user", &UserSchema),
				},
			}},
			"UpdateUser": {ProtoMessageSchema{Name: "UpdateUserResponse", Fields: ProtoFieldsMap{
				1: FieldMask("field_mask"),
			}}, ProtoEmpty()},
		},
	},
	"Post": ProtoServiceSchema{
		Messages: []ProtoMessageSchema{PostSchema},
		Handlers: HandlersMap{
			"GetPost": {GetPostSchema, ProtoMessageSchema{
				Name: "GetPostResponse",
				Fields: ProtoFieldsMap{
					1: MsgField("post", &PostSchema),
				},
			}},
			"UpdatePost": {ProtoMessageSchema{Name: "UpdatePostResponse", Fields: ProtoFieldsMap{
				1: MsgField("post", &PostSchema),
				2: FieldMask("field_mask"),
			}}, ProtoEmpty()},
		},
	},
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
