package schemabuilder

import (
	"log"
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
	Messages:   []ProtoMessageSchema{PostSchema},
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

var UserService = ProtoServiceSchema{
	Resource: UserSchema,
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
	OptionExtensions: OptionExtensions{
		Service: []CustomOption{{"extensionopt", "string", 1, true, true}},
	},
}

var PostService = ProtoServiceSchema{
	Resource: PostSchema,
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
}

var ProtoServices = []ProtoServiceSchema{
	UserService, PostService,
}

var (
	generator = NewProtoGenerator("gen/proto", "myapp.v1")
	services  = BuildServices(ProtoServices)
)

func GenerateProtoFiles() {
	for _, v := range services {
		if err := generator.Generate(v); err != nil {
			log.Fatal(err)
		}
	}
}
