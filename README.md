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
This file gets generated:
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
