# Protoschema hooks

This package contains some premade hooks to use with [protoschema](github.com/Rick-Phoenix/protoschema).

These are the available hooks so far:

## Connect handler generator

This file hook iterates the services in a file and extracts the information from the handlers to automatically generate a full connectRpc handler. It extracts the queries defined in the Queries struct passed to its constructor (which is an instance of the generated sqlc queries) so that *if a query matches the name* of a specific handler (i.e. GetUser), the handler will be generated with that query. This can also be overridden in the handler schema to use custom data instead, in case the names do not match. 

### Requirements

As is the case with [querygen](https://github.com/Rick-Phoenix/querygen), you need to assign these methods to your sqlc Queries instance so that this package can extract the necessary information from the queries in order to use it with file generation:

```go
type QueryData struct {
	Name         string
	ParamName    string
	Params       map[string]string
	ReturnTypes  []string
	ReturnFields map[string]string
	IsResult     bool
	IsErr        bool
	SliceReturn  bool
}

func (q *Queries) GetPkg() string {
	if q == nil {
		return ""
	}

	return reflect.TypeOf(q).Elem().PkgPath()
}

func (q *Queries) ExtractMethods() map[string]*QueryData {
	output := make(map[string]*QueryData)
	model := reflect.TypeOf(q)
	ignoredMethods := []string{"WithTx", "ExtractMethods", "GetPkg"}
	for i := range model.NumMethod() {
		method := model.Method(i)
		data := &QueryData{
			Params:       make(map[string]string),
			ReturnFields: make(map[string]string),
		}
		if slices.Contains(ignoredMethods, method.Name) {
			continue
		}
		data.Name = method.Name
		if method.Type.NumOut() == 1 {
			data.IsErr = true
			data.ReturnTypes = append(data.ReturnTypes, "error")
		} else {
			firstReturn := method.Type.Out(0)

			if firstReturn == reflect.TypeOf((*sql.Result)(nil)).Elem() {
				data.IsResult = true
				data.ReturnTypes = append(data.ReturnTypes, "sql.Result")
			} else {
				var target reflect.Type
				if firstReturn.Kind() == reflect.Slice {
					data.SliceReturn = true
					target = firstReturn.Elem().Elem()
				} else if firstReturn.Kind() == reflect.Pointer {
					target = firstReturn.Elem()
				}

				if target != nil && target.Kind() == reflect.Struct {
					for i := range target.NumField() {
						field := target.Field(i)
						data.ReturnFields[field.Name] = field.Type.Name()
					}
				}

				data.ReturnTypes = append(data.ReturnTypes, target.Name())
				data.ReturnTypes = append(data.ReturnTypes, "error")
			}
		}

		if method.Type.NumIn() > 2 {
			queryParam := method.Type.In(2)
			data.ParamName = queryParam.Name()
			for i := range queryParam.NumField() {
				field := queryParam.Field(i)
				data.Params[field.Name] = field.Type.Name()
			}
		}
		output[data.Name] = data
	}

	return output
}

```

It is also required to use the following sqlc settings, as it makes the parsing of query data much easier:

```yaml
emit_pointers_for_null_types: true
emit_result_struct_pointers: true
query_parameter_limit: 0
```

### Example

Let's look at this configuration:

```go
func TestHandlerGen(t *testing.T) {
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	store := db.New(database)
	handlerBuilder := hooks.NewConnectHandlerGen(store, "gen/handlers")
	files := TestPkg.BuildFiles()
	for _, file := range files {
		err := handlerBuilder.Generate(file)
		assert.NoError(t, err, "Gen handler test")
	}
}
```

As you can see here we are initiating a Queries instance and passing it to the `NewConnectHandlerGen` constructor, as well as specifying the output directory for the generated files. Then we are extracting the FileData with protoschema to pass it to the handler generator. 

From this, this file will be generated:

```go
type UserService struct {
	Queries *db.Queries
}

func NewUserService(queries *db.Queries) *UserService {
	return &UserService{Queries: queries}
}

func (s *UserService) GetUser(
	ctx context.Context,
	req *connect.Request[myappv1.GetUserRequest],
) (*connect.Response[myappv1.GetUserResponse], error) {
	user, err := s.Queries.GetUser(ctx, req.Msg.Get())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		} else {
			var sqliteErr *sqlite.Error
			if errors.As(err, &sqliteErr) {
				fmt.Printf("Sqlite error: %s\n", sqlite.ErrorCodeString[sqliteErr.Code()])
			} else {
				fmt.Printf("Unknown error: %s\n", err.Error())
			}
			return nil, connect.NewError(connect.CodeUnknown, err)
		}
	}

	return connect.NewResponse(&myappv1.GetUserResponse{
		User: converter.UserToUserMsg(user),
	}), nil
}

func (s *UserService) UpdateUser(
	ctx context.Context,
	req *connect.Request[myappv1.UpdateUserRequest],
) (*connect.Response[emptypb.Empty], error) {
	err := s.Queries.UpdateUser(ctx, req.Msg.Get())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		} else {
			var sqliteErr *sqlite.Error
			if errors.As(err, &sqliteErr) {
				fmt.Printf("Sqlite error: %s\n", sqlite.ErrorCodeString[sqliteErr.Code()])
			} else {
				fmt.Printf("Unknown error: %s\n", err.Error())
			}
			return nil, connect.NewError(connect.CodeUnknown, err)
		}
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}
```

>[!NOTE]
> Since this hook is supposed to be plug-and-play with [protoschema](https://github.com/Rick-Phoenix/protoschema), it assumes that you are using the `converter` package that gets generated by it, to convert database structs into their destination protobuf messages type.

Let's recap what is going on here:

Since I have two queries in my Queries struct called GetUser and UpdateUser, the package is automatically extracting those and using them here. As you can see, the UpdateUser is recognized as having only one return value while GetUser is different. By using the methods described above, the package is able to differentiate queries from one another.

>[!NOTE] 
> At the moment the generated handlers include some basic error handling for sqlite queries. In the future I might also implement a postgresql version.

The querydata for a handler can also be specified directly in the handler schema:

```go
type QueryData struct {
	Name         string
	ParamName    string
	Params       map[string]string
	ReturnTypes  []string
	ReturnFields map[string]string
	IsErr        bool
	SliceReturn  bool
	IsResult     bool
}
```
