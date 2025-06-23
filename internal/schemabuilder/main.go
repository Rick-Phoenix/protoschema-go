package schemabuilder

import (
	"errors"
	"fmt"
	"log"
	"maps"
	"os"
	"reflect"
	"slices"
	"strconv"

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
		"name":       ProtoString(1).MinLen(2),
		"posts":      RepeatedField(12, MessageType[gofirst.Post](4, "Post", WithImportPath("myapp/v1/Post.proto"))),
		"created_at": ProtoTimestamp(25),
	},
	ReservedNames:   ReservedNames("name2", "name3"),
	ReservedNumbers: ReservedNumbers(101, 102),
	ReservedRanges:  []Range{{2010, 2029}, {3050, 3055}},
}

var SubRedditSchema = ProtoMessageSchema{
	Name: "Subreddit",
	Fields: ProtoFieldsMap{
		"id":          ProtoInt32(500),
		"name":        ProtoString(1).MinLen(1).MaxLen(48),
		"description": ProtoString(2).MaxLen(255),
		"creator_id":  ProtoInt32(3),
		"posts":       RepeatedField(4, MessageType[gofirst.Post](4, "Post", WithImportPath("myapp/v1/Post.proto"))),
		"created_at":  ProtoTimestamp(5),
	},
}

var PostSchema = ProtoMessageSchema{
	Fields: ProtoFieldsMap{
		"id":           ProtoString(1),
		"created_at":   ProtoTimestamp(2),
		"creator_id":   ProtoInt32(3),
		"title":        ProtoString(4),
		"content":      ProtoString(5),
		"subreddit_id": ProtoInt32(6),
	},
}

type ServicesMap map[string]ProtoServiceSchema

type ServicesData map[string]ProtoService

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

var MyOptions = []CustomOption{{
	Name: "testopt", Type: "string", FieldNr: 1, Optional: true,
}}

var ProtoServices = ServicesMap{
	"User": ProtoServiceSchema{
		Handlers: HandlersMap{
			"GetUser": {PickFields(&UserSchema, "GetUserRequest", "name"), ProtoMessageSchema{
				Name: "GetUserResponse",
				Fields: ProtoFieldsMap{
					"user": MessageType[gofirst.User](1, "User"),
				},
			}},
			"UpdateUser": {ExtendProtoMessage(&UserSchema, ProtoMessageExtension{
				Schema: &ProtoMessageSchema{Name: "UpdateUserResponse", Fields: ProtoFieldsMap{
					"field_mask": FieldMask(210),
				}},
			}), ProtoEmpty()},
		},
	},
	"Post": ProtoServiceSchema{
		Messages: []ProtoMessageSchema{PostSchema},
		Handlers: HandlersMap{
			"GetPost": {PickFields(&PostSchema, "GetPostRequest", "id"), ProtoMessageSchema{
				Name: "GetPostResponse",
				Fields: ProtoFieldsMap{
					"post": MessageType[gofirst.Post](1, "Post"),
				},
			}},
		},
	},
	"Subreddit": ProtoServiceSchema{
		Messages: []ProtoMessageSchema{SubRedditSchema},
		Handlers: HandlersMap{
			"GetSubreddit": {PickFields(&SubRedditSchema, "GetSubredditRequest", "id"), ProtoMessageSchema{
				Name: "GetSubredditResponse",
				Fields: ProtoFieldsMap{
					"subreddit": MessageType[gofirst.Subreddit](1, "Subreddit"),
				},
			}},
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

func CheckDbSchema(model any, schema ProtoFieldsMap, ignores []string) error {
	dbModel := reflect.TypeOf(model).Elem()
	dbModelName := dbModel.Name()
	schemaCopy := maps.Clone(schema)
	var err error

	for i := range dbModel.NumField() {
		dbcol := dbModel.Field(i)
		colname := dbcol.Tag.Get("json")
		ignoreTag := dbcol.Tag.Get("dbignore")
		ignore, _ := strconv.ParseBool(ignoreTag)
		coltype := dbcol.Type.String()

		if pfield, exists := schemaCopy[colname]; exists {
			delete(schemaCopy, colname)
			data, _ := pfield.Build(colname, Set{})
			if data.GoType != coltype && !ignore && !slices.Contains(ignores, data.Name) {
				err = errors.Join(err, fmt.Errorf("Expected type %q for field %q, found %q.", coltype, colname, data.GoType))
			}
		} else if !slices.Contains(ignores, colname) && !ignore {
			err = errors.Join(err, fmt.Errorf("Column %q not found in the proto schema for %q.", colname, dbModel))
		}
	}

	if len(schemaCopy) > 0 {
		for name := range schemaCopy {
			if !slices.Contains(ignores, name) {
				err = errors.Join(err, fmt.Errorf("Unknown field %q found in the schema for db model %q.", name, dbModelName))
			}
		}
	}

	if err != nil {
		err = IndentErrors(fmt.Sprintf("Validation errors for db model %s", dbModelName), err)
	}

	return err
}
