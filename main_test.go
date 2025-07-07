package protoschema_test

import (
	"database/sql"
	"path"
	"testing"

	p "github.com/Rick-Phoenix/protoschema"
	"github.com/Rick-Phoenix/protoschema/hooks"
	"github.com/Rick-Phoenix/protoschema/test/db"
	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
)

var (
	goMod   = "github.com/Rick-Phoenix/protoschema"
	TestPkg = p.NewProtoPackage(p.ProtoPackageConfig{
		Name:      "myapp.v1",
		ProtoRoot: "testdata/proto",
		GoPackage: path.Join(goMod, "gen/myappv1"),
		GoModule:  goMod,
	})
)

var PostFile = TestPkg.NewFile(p.FileSchema{
	Package: TestPkg,
	Name:    "post",
})

var PostService = PostFile.NewService(p.ServiceSchema{
	Resource: "Post",
	Handlers: p.HandlersMap{
		"GetPost": {
			Request:  GetPostRequest,
			Response: GetPostResponse,
		},
		"UpdatePost": {
			Request:  UpdatePostRequest,
			Response: p.Empty(),
		},
	},
})

var PostSchema = PostFile.NewMessage(p.MessageSchema{
	Name: "Post",
	Fields: p.FieldsMap{
		1: p.Int64("id"),
		2: p.Timestamp("created_at"),
		3: p.Int64("author_id"),
		4: p.String("title").MinLen(5).MaxLen(64).Required(),
		5: p.String("content").Optional(),
		6: p.Int64("subreddit_id"),
	},
	ModelIgnore: []string{"content", "updated_at"},
	Model:       &db.Post{},
})

var GetPostRequest = PostFile.NewMessage(p.MessageSchema{
	Name: "GetPostRequest",
	Fields: p.FieldsMap{
		1: PostSchema.GetField("id"),
	},
})

var GetPostResponse = PostFile.NewMessage(p.MessageSchema{
	Name: "GetPostResponse",
	Fields: p.FieldsMap{
		1: p.MsgField("post", PostSchema),
	},
})

var UpdatePostRequest = PostFile.NewMessage(p.MessageSchema{Name: "UpdatePostRequest", Fields: p.FieldsMap{
	1: p.MsgField("post", PostSchema),
	2: p.FieldMask("field_mask"),
}})

var UserFile = TestPkg.NewFile(p.FileSchema{
	Name:    "user",
	Package: TestPkg,
})

var UserSchema = UserFile.NewMessage(p.MessageSchema{
	Name: "User",
	Fields: p.FieldsMap{
		1: p.String("name").Required().MinLen(2).MaxLen(32),
		2: p.Int64("id"),
		3: p.Timestamp("created_at"),
		4: p.Repeated("posts", p.MsgField("post", PostSchema)),
	},
	Model: &db.UserWithPosts{},
})

var GetUserRequest = UserFile.NewMessage(p.MessageSchema{
	Name: "GetUserRequest",
	Fields: p.FieldsMap{
		1: p.Int64("id").Required(),
	},
})

var GetUserResponse = UserFile.NewMessage(p.MessageSchema{
	Name: "GetUserResponse",
	Fields: p.FieldsMap{
		1: p.MsgField("user", UserSchema),
	},
})

var UpdateUserRequest = UserFile.NewMessage(p.MessageSchema{
	Name: "UpdateUserRequest",
	Fields: p.FieldsMap{
		1: p.Int64("id").Required(),
		2: p.FieldMask("field_mask"),
	},
})

var UserService = UserFile.NewService(p.ServiceSchema{
	Resource: "User",
	Handlers: p.HandlersMap{
		"GetUser": {
			Request:  GetUserRequest,
			Response: GetUserResponse,
		},
		"UpdateUser": {
			Request:  UpdateUserRequest,
			Response: p.Empty(),
		},
	},
})

func TestMain(t *testing.T) {
	// assert.NoError(t, err, "Main Test")
}

func TestHandlerGen(t *testing.T) {
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	store := db.New(database)
	handlerBuilder := hooks.NewHandlerBuilder(store, "gen/handlers")
	files := TestPkg.BuildFiles()
	for _, file := range files {
		err := handlerBuilder.Generate(file)
		assert.NoError(t, err, "Gen handler test")
	}
}
