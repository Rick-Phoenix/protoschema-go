package schemabuilder

import (
	"path"
	"testing"
	"time"

	"github.com/Rick-Phoenix/gofirst/db"
	"github.com/Rick-Phoenix/gofirst/db/sqlgen"
	"github.com/Rick-Phoenix/gofirst/gen/myappv1"
	"github.com/stretchr/testify/assert"
)

var (
	goMod        = "github.com/Rick-Phoenix/gofirst"
	protoPackage = NewProtoPackage(ProtoPackageConfig{
		Name:      "myapp.v1",
		ProtoRoot: "proto",
		GoPackage: path.Join(goMod, "gen/myappv1"),
		GoModule:  goMod,
	})
)

var SubRedditSchema = protoPackage.NewMessage(MessageSchema{
	Name: "Subreddit",
	Fields: FieldsMap{
		1: Int32("id").Optional(),
		2: String("name").MinLen(1).MaxLen(48),
		3: String("description").MaxLen(255),
		4: Int32("creator_id"),
		5: Repeated("posts", MsgField("post", &PostSchema)),
		6: Timestamp("created_at"),
	},
})

var PostSchema = protoPackage.NewMessage(MessageSchema{
	Name: "Post",
	Fields: FieldsMap{
		1: Int64("id"),
		2: Timestamp("created_at"),
		3: Int64("author_id"),
		4: String("title").MinLen(5).MaxLen(64).Required(),
		5: String("content").Optional(),
		6: Int64("subreddit_id"),
	},
	ImportPath:  "myapp/v1/post.proto",
	ModelIgnore: []string{"content"},
	Model:       &sqlgen.Post{},
	TargetType:  &myappv1.Post{},
})

var GetPostRequest = protoPackage.NewMessage(MessageSchema{
	Name: "GetPostRequest",
	Fields: FieldsMap{
		1: PostSchema.GetField("id"),
	},
})

var GetPostResponse = protoPackage.NewMessage(MessageSchema{
	Name: "GetPostResponse",
	Fields: FieldsMap{
		1: MsgField("post", &PostSchema),
	},
})

var UpdatePostRequest = protoPackage.NewMessage(MessageSchema{Name: "UpdatePostRequest", Fields: FieldsMap{
	1: MsgField("post", &PostSchema),
	2: FieldMask("field_mask"),
}})

var PostService = ServiceSchema{
	Resource: PostSchema,
	Handlers: HandlersMap{
		"GetPost": {
			GetPostRequest,
			GetPostResponse,
		},
		"UpdatePost": {
			UpdatePostRequest,
			Empty(),
		},
	},
}

type UserWithPosts struct {
	Id        int64         `json:"id"`
	Name      string        `json:"name"`
	CreatedAt time.Time     `json:"created_at"`
	Posts     []sqlgen.Post `json:"posts"`
}

var UserSchema = protoPackage.NewMessage(MessageSchema{
	Name: "User",
	Fields: FieldsMap{
		1: String("name").Required().MinLen(2).MaxLen(32),
		2: Int64("id"),
		3: Timestamp("created_at"),
		4: Repeated("posts", MsgField("post", &PostSchema)),
	},
	Model:      &db.UserWithPosts{},
	ImportPath: "myapp/v1/user.proto",
	TargetType: "myappv1.User",
})

var GetUserRequest = protoPackage.NewMessage(MessageSchema{
	Name: "GetUserRequest",
	Fields: FieldsMap{
		1: Int64("id").Required(),
	},
})

var GetUserResponse = protoPackage.NewMessage(MessageSchema{
	Name: "GetUserResponse",
	Fields: FieldsMap{
		1: MsgField("user", &UserSchema),
	},
})

var UpdateUserRequest = protoPackage.NewMessage(MessageSchema{
	Name: "UpdateUserRequest",
	Fields: FieldsMap{
		1: Int64("id").Required(),
		2: FieldMask("field_mask"),
	},
})

var UserService = ServiceSchema{
	Resource: UserSchema,
	Handlers: HandlersMap{
		"GetUser": {
			GetUserRequest, GetUserResponse,
		},
		"UpdateUser": {
			UpdateUserRequest,
			Empty(),
		},
	},
}

func TestMain(t *testing.T) {
	generator := protoPackage.Services(UserService, PostService)
	err := generator.Generate()
	assert.NoError(t, err, "Main Test")
}
