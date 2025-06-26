package schemabuilder

import (
	"log"
	"testing"
	"time"

	gofirst "github.com/Rick-Phoenix/gofirst/db/queries/gen"
)

type UserWithPosts struct {
	ID        int64          `json:"id"`
	Name      string         `json:"name"`
	CreatedAt time.Time      `dbignore:"true" json:"created_at"`
	Posts     []gofirst.Post `json:"posts"`
}

var UserSchema = MessageSchema{
	Name: "User",
	Fields: FieldsMap{
		1: String("name"),
		2: Int64("id"),
		3: Timestamp("created_at"),
		5: Repeated("posts", MsgField("post", &PostSchema)).CelOption("myexpr", "x must not be y", "x != y"),
	},
	Enums: []EnumGroup{
		{
			Name: "myenum",
			Members: EnumMembers{
				0: "VAL_1",
				1: "VAL_2",
			},
			Options: []ProtoOption{{"myopt1", "myval1"}, {"myopt", "myval"}},
		},
	},
	Model:      &UserWithPosts{},
	ImportPath: "myapp/v1/user.proto",
	Options:    []ProtoOption{MessageCelOption("myexpr", "x must not be y", "x != y")},
	Messages:   []MessageSchema{PostSchema},
}

var GetUserSchema = MessageSchema{
	Name: "GetUserRequest",
	Fields: FieldsMap{
		1: UserSchema.GetField("name"),
	},
}

var GetPostSchema = MessageSchema{
	Name: "GetPostRequest",
	Fields: FieldsMap{
		1: PostSchema.GetField("id"),
	},
}

var SubRedditSchema = MessageSchema{
	Name: "Subreddit",
	Fields: FieldsMap{
		1: Int32("id").Optional(),
		2: String("name").MinLen(1).MaxLen(48),
		3: String("description").MaxLen(255),
		4: Int32("creator_id"),
		5: Repeated("posts", MsgField("post", &PostSchema)),
		6: Timestamp("created_at"),
	},
}

var PostSchema = MessageSchema{
	Name: "Post",
	Fields: FieldsMap{
		1: Int64("id").Optional(),
		2: Timestamp("created_at"),
		3: Int64("author_id"),
		4: String("title").MinLen(5).MaxLen(64).Required(),
		5: String("content").Optional(),
		6: Int64("subreddit_id"),
	},
	ImportPath:  "myapp/v1/post.proto",
	ModelIgnore: []string{"content"},
	Model:       &gofirst.Post{},
}

var UserService = ServiceSchema{
	Resource: UserSchema,
	Handlers: HandlersMap{
		"GetUser": {GetUserSchema, MessageSchema{
			Name: "GetUserResponse",
			Fields: FieldsMap{
				1: MsgField("user", &UserSchema),
			},
		}},
		"UpdateUser": {MessageSchema{Name: "UpdateUserRequest", Fields: FieldsMap{
			1: FieldMask("field_mask"),
			2: MsgField("user", &UserSchema),
		}}, Empty()},
	},
	FileOptions:    []ProtoOption{{"myopt1", "myval1"}, {"myopt", "myval"}},
	ServiceOptions: []ProtoOption{{"myopt1", "myval1"}, {"myopt", "myval"}},
	OptionExtensions: Extensions{
		Service: []ExtensionField{{"extensionopt", "string", 1, true, true}},
	},
}

var PostService = ServiceSchema{
	Resource: PostSchema,
	Handlers: HandlersMap{
		"GetPost": {GetPostSchema, MessageSchema{
			Name: "GetPostResponse",
			Fields: FieldsMap{
				1: MsgField("post", &PostSchema),
			},
		}},
		"UpdatePost": {MessageSchema{Name: "UpdatePostRequest", Fields: FieldsMap{
			1: MsgField("post", &PostSchema),
			2: FieldMask("field_mask"),
		}}, Empty()},
	},
}

var ProtoServices = []ServiceSchema{
	UserService, PostService,
}

var (
	generator = NewProtoGenerator("gen/proto", "myapp.v1")
	services  = BuildServices(ProtoServices)
)

func TestMain(t *testing.T) {
	for _, v := range services {
		if err := generator.Generate(v); err != nil {
			log.Fatal(err)
		}
	}
}
