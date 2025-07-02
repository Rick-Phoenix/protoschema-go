# The goal of this package

protoschema aims to do the following:

1. Define the contents of protobuf files in a declarative way, directly from go code
2. Provide a simple, zod-inspired api to easily add protovalidate rules (and other kinds of options) to fields or messages
3. Couple messages to specific models (like items coming from a database or from another api), and report errors if there is mismatch between the two
4. Automatically generate functions to convert source types (like the one above) to the go types generated from the protobuf messages definition
5. Expose various kinds of hooks that can be used to read the data for each field, message or file and perform custom actions like generating other files based on a customized template (or to perform additional validation)

# What it does

## Proto files generation

From this schema declaration:

```go
// Define the package
var (
	goMod        = "github.com/Rick-Phoenix/protoschema"
	protoPackage = NewProtoPackage(ProtoPackageConfig{
		Name:      "myapp.v1",
		ProtoRoot: "proto",
		GoPackage: path.Join(goMod, "gen/myappv1"),
		GoModule:  goMod,
	})
)

// Add file to package
var PostFile = protoPackage.NewFile(FileSchema{
	Package: protoPackage,
	Name:    "post",
})

// Add service to file
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

// Add message to file, with model for validation
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
    // Call the generate function
	err := protoPackage.Generate()
	assert.NoError(t, err, "Main Test")
}
```

Where the models being used are these:

```go
// sqlgen package (generated with sqlc)
type Post struct {
	Id          int64     `json:"id"`
	Title       string    `json:"title"`
	Content     *string   `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	AuthorId    int64     `json:"author_id"`
	SubredditId int64     `json:"subreddit_id"`
}

type User struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// db package
type UserWithPosts struct {
	sqlgen.User
	Posts []*sqlgen.Post
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

>[!NOTE]
> protoschema uses the buf cli with the `buf format` command to prettify the output of the code generation. It is highly encouraged to download the buf cli and make it available in path to avoid having messy-looking proto files.

## Converter functions

The package will also generate some functions that can be used to easily convert a struct from its original model type (usually a database item) to the message type that will be used in responses. 

At the moment, the converters generator can only handle the following field types:

- Basic types (string, int, etc) 
- Types that refer to other message schemas belonging to the same package 
- time.Time, which gets converted with timestamppb.New. 

However, the package also allows the user to define their custom function that receives the data for the field (along with the context of its file, package, message and message model) and overrides the default function. 
(Function signature is shown below)

This is what the converter package would look like for the schema above:

```go
package converter

import (
	"github.com/Rick-Phoenix/protoschema/db"
	"github.com/Rick-Phoenix/protoschema/db/sqlgen"
	"github.com/Rick-Phoenix/protoschema/gen/myappv1"
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

// If a message is used as repeated in a field, a slice converter like this will also be generated 
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

Converter function that can be used to replace the default variant:

```go
type ConverterFunc func(ConverterFuncData)

type ConverterFuncData struct {
	Package    *ProtoPackage
	File       *FileSchema
	Message    *MessageSchema
	ModelField reflect.StructField
	ProtoField FieldBuilder
}
```

This function will be called for each model field being iterated during the validation step, and receive all the data for that field, along with the surrounding Package, File and Message.

## Model validation

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
    // Now a string
	Id           string    `json:"id"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
    // Extra field
	ExtraDbField string    `json:"extra_db_field"`
}
```

While also adding an extra field to the message schema which is not present in its model:

```go
5: String("non_db_field")
```

This will cause the package to report these errors and exit:

```
❌ The following errors occurred:
Errors in the file user.proto:
  Errors for the User message schema:
    Validation errors for model db.UserWithPosts:
      Expected type "string" for field "id", found "int64".
      Model field "extra_db_field" not found in the message schema.
      Unknown field "non_db_field" is not present in the message's model.
```

This ensures that if a change occurs on either side but is not implemented on the other side, the proto files will not be generated (unless the user specifically chooses to skip validation for a given field or for an entire message).

## Hooks

protoschema also allows the user to define hooks for the whole package or for single schemas, which will be called when the schema is processed, and receive all the data for that protobuf element (file, service, oneof or message), which can be used to perform custom actions such as code generation. 

```go
// Also available for messages, services and oneofs
type FileHook func(d FileData) error

type FileData struct {
	Package    *ProtoPackage
	Name       string
	Imports    Set
	Extensions Extensions
	Options    []ProtoOption
	Enums      []EnumGroup
	Messages   []MessageData
	Services   []ServiceData
	Metadata   map[string]any
}
```

# State of the project

protoschema is still in the beta stage. It may receive some breaking changes in the near future.

# Tools being used

- [Protocompile](https://github.com/bufbuild/protocompile) (parser and reporter, used to extract the data from proto files in tests) 

# Inspirations

This project was inspired by:
- [Zod](https://github.com/colinhacks/zod), for its beautiful, ergonomic api for defining validation rules
- [protovalidate](https://github.com/bufbuild/protovalidate), for its innovative approach of defining validation rules directly within proto files
