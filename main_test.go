package schemabuilder

import (
	"testing"
	"time"

	"github.com/Rick-Phoenix/gofirst/db"
	"github.com/Rick-Phoenix/gofirst/db/sqlgen"
	"github.com/stretchr/testify/assert"
)

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
}

var GetPostRequest = MessageSchema{
	Name: "GetPostRequest",
	Fields: FieldsMap{
		1: PostSchema.GetField("id"),
	},
}

var GetPostResponse = MessageSchema{
	Name: "GetPostResponse",
	Fields: FieldsMap{
		1: MsgField("post", &PostSchema),
	},
}

var UpdatePostRequest = MessageSchema{Name: "UpdatePostRequest", Fields: FieldsMap{
	1: MsgField("post", &PostSchema),
	2: FieldMask("field_mask"),
}}

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
	ID        int64         `json:"id"`
	Name      string        `json:"name"`
	CreatedAt time.Time     `dbignore:"true" json:"created_at"`
	Posts     []sqlgen.Post `json:"posts"`
}

var UserSchema = MessageSchema{
	Name: "User",
	Fields: FieldsMap{
		1: String("name").Required().MinLen(2).MaxLen(32),
		2: Int64("id"),
		3: Timestamp("created_at"),
		4: Repeated("posts", MsgField("post", &PostSchema)),
	},
	Model:      &db.UserWithPosts{},
	ImportPath: "myapp/v1/user.proto",
}

var GetUserRequest = MessageSchema{
	Name: "GetUserRequest",
	Fields: FieldsMap{
		1: UserSchema.GetField("id"),
	},
}

var GetUserResponse = MessageSchema{
	Name: "GetUserResponse",
	Fields: FieldsMap{
		1: MsgField("user", &UserSchema),
	},
}

var UpdateUserRequest = MessageSchema{Name: "UpdateUserRequest", Fields: FieldsMap{
	1: MsgField("user", &UserSchema),
	2: FieldMask("field_mask"),
}}

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
	generator := NewProtoGenerator("proto", "myapp.v1").Services(UserService, PostService)
	err := generator.Generate()
	assert.NoError(t, err, "Main Test")
}
