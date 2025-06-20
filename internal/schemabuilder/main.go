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
		"maptype":    ProtoMap(209, ProtoInt32(0).Lt(10), ProtoEnumField(100, "Myenum").DefinedOnly().In(1, 2, 3)),
		"maptype2":   ProtoMap(201, ProtoInt32(0).Lt(10), ProtoTimestamp(25).LtNow()),
		"created_at": ProtoTimestamp(25).LtNow(),
		"enumval":    ProtoEnumField(100, "Myenum").DefinedOnly().In(1, 2, 3).Optional(),
		"repenum":    RepeatedField(125, ProtoEnumField(100, "Myenum").DefinedOnly().In(1, 2, 3)),
	},
}

var PostSchema = ProtoMessageSchema{
	Fields: ProtoFieldsMap{
		"id": ProtoString(1),
		"createdAt": ProtoTimestamp(2).CelOptions(CelFieldOpts{
			Id:         "test",
			Message:    "this is a test",
			Expression: "this = test",
		}).LtNow(),
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

var TablesData = ServicesMap{
	"User": ProtoServiceSchema{
		Handlers: HandlersMap{
			"GetUser":    {ProtoEmpty(), UserSchema},
			"UpdateUser": {MessageRef("UpdateUserResponse"), ProtoEmpty()},
		},
		Enums: []ProtoEnumGroup{ProtoEnum("Myenum", ProtoEnumMap{"VAL_1": 0, "VAL_2": 1}).SetReservedNumbers(20, 200).SetReservedNames("RESERVED_NAME")},
	},
	"Post": ProtoServiceSchema{
		Messages: []ProtoMessageSchema{PostSchema},
		Handlers: HandlersMap{
			"GetPost": {PostSchema, PostSchema},
		},
	},
}

func GenerateProtoFiles() {
	var Services = BuildFinalServicesMap(TablesData)
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
