From this input:
```go
var (
	goMod        = "github.com/Rick-Phoenix/gofirst"
	protoPackage = NewProtoPackage(ProtoPackageConfig{
		Name:      "myapp.v1",
		ProtoRoot: "proto",
		GoPackage: path.Join(goMod, "gen/myappv1"),
		GoModule:  goMod,
	})
)

var PostFile = protoPackage.NewFile(FileSchema{
	Package: protoPackage,
	Name:    "post",
})

var PostService = PostFile.NewService(ServiceSchema{
	Resource: "Post",
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
})

var PostSchema = PostFile.NewMessage(MessageSchema{
	Name: "Post",
	Fields: FieldsMap{
		1: Int64("id"),
		2: Timestamp("created_at"),
		3: Int64("author_id"),
		4: String("title").MinLen(5).MaxLen(64).Required(),
		5: String("content").Optional(),
		6: Int64("subreddit_id"),
	},
	ModelIgnore: []string{"content"},
	Model:       &sqlgen.Post{},
})

var GetPostRequest = PostFile.NewMessage(MessageSchema{
	Name: "GetPostRequest",
	Fields: FieldsMap{
		1: PostSchema.GetField("id"),
	},
})

var GetPostResponse = PostFile.NewMessage(MessageSchema{
	Name: "GetPostResponse",
	Fields: FieldsMap{
		1: MsgField("post", PostSchema),
	},
})

var UpdatePostRequest = PostFile.NewMessage(MessageSchema{Name: "UpdatePostRequest", Fields: FieldsMap{
	1: MsgField("post", PostSchema),
	2: FieldMask("field_mask"),
}})

type UserWithPosts struct {
	Id        int64         `json:"id"`
	Name      string        `json:"name"`
	CreatedAt time.Time     `json:"created_at"`
	Posts     []sqlgen.Post `json:"posts"`
}

var UserFile = protoPackage.NewFile(FileSchema{
	Name:    "user",
	Package: protoPackage,
	Hook:    protoPackage.genConnectHandler,
})

var UserSchema = UserFile.NewMessage(MessageSchema{
	Name: "User",
	Fields: FieldsMap{
		1: String("name").Required().MinLen(2).MaxLen(32),
		2: Int64("id"),
		3: Timestamp("created_at"),
		4: Repeated("posts", MsgField("post", PostSchema)),
	},
	Model: &db.UserWithPosts{},
})

var GetUserRequest = UserFile.NewMessage(MessageSchema{
	Name: "GetUserRequest",
	Fields: FieldsMap{
		1: Int64("id").Required(),
	},
})

var GetUserResponse = UserFile.NewMessage(MessageSchema{
	Name: "GetUserResponse",
	Fields: FieldsMap{
		1: MsgField("user", UserSchema),
	},
})

var UpdateUserRequest = UserFile.NewMessage(MessageSchema{
	Name: "UpdateUserRequest",
	Fields: FieldsMap{
		1: Int64("id").Required(),
		2: FieldMask("field_mask"),
	},
})

var UserService = UserFile.NewService(ServiceSchema{
	Resource: "User",
	Handlers: HandlersMap{
		"GetUser": {
			GetUserRequest, GetUserResponse,
		},
		"UpdateUser": {
			UpdateUserRequest,
			Empty(),
		},
	},
})

func TestMain(t *testing.T) {
	err := protoPackage.Generate()
	assert.NoError(t, err, "Main Test")
}
```
These two files are generated:

- user.proto
```proto
syntax = "proto3";

package myapp.v1;

import "buf/validate/validate.proto";
import "google/protobuf/empty.proto";
import "google/protobuf/field_mask.proto";
import "google/protobuf/timestamp.proto";
import "myapp/v1/post.proto";

message User {
  string name = 1 [
    (buf.validate.field).required = true,
    (buf.validate.field).string.max_len = 32,
    (buf.validate.field).string.min_len = 2
  ];
  int64 id = 2;
  google.protobuf.Timestamp created_at = 3;
  repeated Post posts = 4;
}

message GetUserRequest {
  int64 id = 1 [(buf.validate.field).required = true];
}

message GetUserResponse {
  User user = 1;
}

message UpdateUserRequest {
  int64 id = 1 [(buf.validate.field).required = true];
  google.protobuf.FieldMask field_mask = 2;
}

service UserService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
  rpc UpdateUser(UpdateUserRequest) returns (google.protobuf.Empty);
}
```

- post.proto

```proto
syntax = "proto3";

package myapp.v1;

import "buf/validate/validate.proto";
import "google/protobuf/empty.proto";
import "google/protobuf/field_mask.proto";
import "google/protobuf/timestamp.proto";

message Post {
  int64 id = 1;
  google.protobuf.Timestamp created_at = 2;
  int64 author_id = 3;
  string title = 4 [
    (buf.validate.field).required = true,
    (buf.validate.field).string.max_len = 64,
    (buf.validate.field).string.min_len = 5
  ];
  optional string content = 5;
  int64 subreddit_id = 6;
}

message GetPostRequest {
  int64 id = 1;
}

message GetPostResponse {
  Post post = 1;
}

message UpdatePostRequest {
  Post post = 1;
  google.protobuf.FieldMask field_mask = 2;
}

service PostService {
  rpc GetPost(GetPostRequest) returns (GetPostResponse);
  rpc UpdatePost(UpdatePostRequest) returns (google.protobuf.Empty);
}
```

Along with some converter functions:

```go
package converter

import (
	"github.com/Rick-Phoenix/gofirst/db"
	"github.com/Rick-Phoenix/gofirst/db/sqlgen"
	"github.com/Rick-Phoenix/gofirst/gen/myappv1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func PostToPostMsg(Post *sqlgen.Post) *myappv1.Post {
	if Post == nil {
		return nil
	}
	CreatedAt := timestamppb.New(Post.CreatedAt)
	return &myappv1.Post{
		Id:          Post.Id,
		Title:       Post.Title,
		Content:     Post.Content,
		CreatedAt:   CreatedAt,
		AuthorId:    Post.AuthorId,
		SubredditId: Post.SubredditId,
	}
}

func PostsToPostsMsg(Post []*sqlgen.Post) []*myappv1.Post {
	out := make([]*myappv1.Post, len(Post))

	for i, v := range Post {
		out[i] = PostToPostMsg(v)
	}

	return out
}
func UserToUserMsg(User *db.UserWithPosts) *myappv1.User {
	if User == nil {
		return nil
	}
	CreatedAt := timestamppb.New(User.CreatedAt)
	return &myappv1.User{
		Id:        User.Id,
		Name:      User.Name,
		CreatedAt: CreatedAt,
		Posts:     PostsToPostsMsg(User.Posts),
	}
}
```

At the moment, the converters generator can only handle basic types, types that come from other message schemas, or time.Time types, which are converted with timestamppb.New. 

The package also allows the user to define their custom function that receives the data for the field (along with the context of its file, package, message and message model) and overrides the default function.

The package handles validation for message schemas. When a message schema has a defined model, (like the `&db.UserWithPosts{}`), the package will show an error if the types in the schema do not match the types in the model struct, or if a field is present in one but not in the other (a ModelIgnore slice of strings can be used to ignore specific fields if necessary).

For example, let's try changing the User model from this:

```go
type User struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
```

To this:

```go
type User struct {
	Id           string    `json:"id"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	ExtraDbField string    `json:"extra_db_field"`
}
```

While also adding an extra field to the message schema which is not present in its model:

`5: String("non_db_field")`

This will cause protovalidate to report these errors and exit:

`  ❌ The following errors occurred:
Errors in the file user.proto:
  Errors for the User message schema:
    Validation errors for model db.UserWithPosts:
      Expected type "string" for field "id", found "int64".
      Column "extra_db_field" not found in the proto schema for "sqlgen.User".
      Unknown field "non_db_field" found in the message schema for model "db.UserWithPosts".`

This ensures that if a change occurs on either side but is not implemented on the other side, the proto files will not be generated (unless the user specifically chooses to skip validation for a given field or for an entire message).
