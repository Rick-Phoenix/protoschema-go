package schemabuilder_test

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	sb "github.com/Rick-Phoenix/gofirst"
	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FieldData struct {
	Name            string
	Number          int32
	Repeated        bool
	Optional        bool
	TypeName        string
	Options         map[string]ProtoOption
	RepeatedOptions []ProtoOption
}

type EnumMember struct {
	Number int32
}

type EnumData struct {
	Name            string
	Members         map[string]EnumMember
	Options         map[string]ProtoOption
	RepeatedOptions []ProtoOption
	ReservedNames   []string
	ReservedNumbers []int32
	ReservedRanges  []sb.Range
}

type ProtoOption struct {
	Name  string
	Value any
}

type OneofData struct {
	Name            string
	Options         map[string]ProtoOption
	RepeatedOptions []ProtoOption
}

type MessageData struct {
	Name            string
	Fields          map[string]FieldData
	Enums           map[string]EnumData
	Oneofs          map[string]OneofData
	Messages        map[string]MessageData
	ReservedNames   []string
	ReservedNumbers []int32
	ReservedRanges  []sb.Range
	Options         map[string]ProtoOption
	RepeatedOptions []ProtoOption
}

type ServiceData struct {
	Name            string
	Handlers        map[string]MethodData
	Options         map[string]ProtoOption
	RepeatedOptions []ProtoOption
}

type MethodData struct {
	Name       string
	InputType  string
	OutputType string
}

type FileData struct {
	Package    string
	Imports    []string
	Messages   map[string]MessageData
	Enums      map[string]EnumData
	Services   map[string]ServiceData
	Extensions map[string]ExtensionData
}

type ExtensionData struct {
	Extendee string
	Fields   map[string]FieldData
}

type UserWithPosts struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `dbignore:"true" json:"created_at"`
	Posts     []Post    `json:"posts"`
}

type Post struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Content     *string   `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	AuthorID    int64     `json:"author_id"`
	SubredditID int64     `json:"subreddit_id"`
}

func TestGeneration(t *testing.T) {
	var (
		timePast   = timestamppb.Timestamp{Seconds: time.Date(2025, time.January, 1, 1, 1, 1, 1, time.Local).Unix()}
		timeFuture = timestamppb.Timestamp{Seconds: time.Date(3000, time.January, 1, 1, 1, 1, 1, time.Local).Unix()}
	)

	MyOptions := []sb.ExtensionField{{
		Name: "testopt", Type: "string", FieldNr: 1, Optional: true,
	}}
	TestExtensions := sb.Extensions{
		Message: MyOptions,
		File:    MyOptions,
		OneOf:   MyOptions,
		Service: MyOptions,
		Field:   MyOptions,
	}

	// tmpDir := t.TempDir()
	tmpDir := "gen/temp"
	goMod := "github.com/Rick-Phoenix/gofirst"
	config := sb.ProtoPackageConfig{
		Name:               "myapp.v1",
		GoModule:           goMod,
		GoPackage:          path.Join(goMod, "gen/myappv1"),
		ProtoRoot:          tmpDir,
		ConverterOutputDir: filepath.Join(tmpDir, "converter"),
	}
	gen := sb.NewProtoPackage(config)

	PostFile := gen.NewFile(sb.FileSchema{
		Name: "post",
	})

	PostSchema := PostFile.NewMessage(sb.MessageSchema{
		Name: "Post",
		Fields: sb.FieldsMap{
			1: sb.Int64("id").Optional(),
			2: sb.Timestamp("created_at"),
			3: sb.Int64("author_id"),
			4: sb.String("title").MinLen(5).MaxLen(64).Required(),
			5: sb.String("content").Optional().Nullable(),
			6: sb.Int64("subreddit_id"),
		},
		Model: &Post{},
	})

	UserFile := gen.NewFile(sb.FileSchema{
		Name:       "user",
		Extensions: TestExtensions,
	})

	UserSchema := UserFile.NewMessage(sb.MessageSchema{
		Name: "User",
		Fields: sb.FieldsMap{
			1: sb.String("name").Required().MinLen(2).MaxLen(32),
			2: sb.Int64("id"),
			3: sb.Timestamp("created_at"),
			4: sb.Repeated("posts", sb.MsgField("post", PostSchema)),
			5: sb.String("fav_cat").
				Optional().
				CelOption("cel", "msg", "expr").
				CelOption("cel", "msg", "expr").
				Options([]sb.ProtoOption{{Name: "myopt", Value: true}, {Name: "myopt", Value: false}}...).
				RepeatedOptions([]sb.ProtoOption{{Name: "repopt", Value: true}, {Name: "repopt", Value: true}}...).
				Example("tabby").
				Example("calico"),
			6: sb.Map("mymap", sb.String("").MinLen(1), sb.Int64("").Gt(1).In(1, 2)).MinPairs(2).MaxPairs(4),
			7: sb.Repeated("reptest", sb.Int32("").Gt(1).In(1, 2)).Unique().MinItems(1).MaxItems(4),
			8: sb.Timestamp("timetest").Lt(&timePast),
			9: sb.Timestamp("timetest2").Const(&timePast),
		},
		ReservedNames:   []string{"name1", "name2"},
		ReservedNumbers: []uint{20, 21},
		ReservedRanges:  []sb.Range{{22, 25}},
	})

	UserSchema.NewOneof(sb.OneofGroup{
		Name: "myoneof",
		Fields: sb.OneofFields{
			9:  sb.String("example"),
			10: sb.Int32("another"),
		},
	})

	TestEnum := sb.EnumGroup{
		Name: "myenum",
		Members: sb.EnumMembers{
			0: "VAL_1",
			1: "VAL_2",
		},
		Options:         []sb.ProtoOption{sb.Options.AllowAlias},
		ReservedNames:   []string{"name1", "name2"},
		ReservedNumbers: []int32{10, 11},
		ReservedRanges:  []sb.Range{{20, 23}},
	}

	UserFile.NewEnum(TestEnum)
	UserSchema.NewEnum(TestEnum)

	NestedUserMsg := UserSchema.NestedMessage(sb.MessageSchema{
		Name: "nested",
		Fields: sb.FieldsMap{
			1: sb.String("example"),
		},
	})

	UserTestEnum := NestedUserMsg.NewEnum(TestEnum)

	GetUserRequest := UserFile.NewMessage(sb.MessageSchema{
		Name: "GetUserRequest", Fields: sb.FieldsMap{
			1: sb.Int64("id"),
			2: sb.EnumField("nestedenum", UserTestEnum),
		},
	})

	GetUserResponse := UserFile.NewMessage(sb.MessageSchema{
		Name: "GetUserResponse",
		Fields: sb.FieldsMap{
			1: sb.MsgField("user", UserSchema),
		},
	})

	UpdateUserRequest := UserFile.NewMessage(sb.MessageSchema{
		Name: "UpdateUserRequest",
		Fields: sb.FieldsMap{
			1: sb.FieldMask("field_mask"),
			2: sb.MsgField("user", UserSchema),
		},
	})

	UserFile.NewService(sb.ServiceSchema{
		Resource: "User",
		Handlers: sb.HandlersMap{
			"GetUser": {
				GetUserRequest,
				GetUserResponse,
			},
			"UpdateUser": {
				UpdateUserRequest,
				sb.Empty(),
			},
		},
	})

	err := gen.Generate()
	if err != nil {
		log.Fatal(err)
	}

	filePath := path.Join(tmpDir, "myapp/v1", "user.proto")

	out := ParseProtoFile(filePath)

	userMsg := out.Messages["User"]

	equalTests := []struct {
		Target   any
		Expected any
	}{
		{out.Package, "myapp.v1"},
		{out.Messages["User"].Name, "User"},
		{out.Services["UserService"].Handlers["GetUser"].InputType, "GetUserRequest"},
		{out.Services["UserService"].Handlers["GetUser"].OutputType, "GetUserResponse"},
		{out.Services["UserService"].Handlers["UpdateUser"].InputType, "UpdateUserRequest"},
		{out.Services["UserService"].Handlers["UpdateUser"].OutputType, "google.protobuf.Empty"},
		{out.Extensions["google.protobuf.MessageOptions"].Fields["testopt"].Number, int32(1)},
		{out.Extensions["google.protobuf.MessageOptions"].Fields["testopt"].Optional, true},
		{out.Extensions["google.protobuf.FileOptions"].Fields["testopt"].Number, int32(1)},
		{out.Extensions["google.protobuf.ServiceOptions"].Fields["testopt"].Number, int32(1)},
		{out.Extensions["google.protobuf.OneofOptions"].Fields["testopt"].Number, int32(1)},
		{out.Extensions["google.protobuf.FieldOptions"].Fields["testopt"].Number, int32(1)},
		{out.Enums["myenum"].Members["VAL_1"].Number, int32(0)},
		{out.Enums["myenum"].Members["VAL_2"].Number, int32(1)},
		{out.Enums["myenum"].Options["allow_alias"].Value, true},
		{userMsg.Enums["myenum"].Members["VAL_1"].Number, int32(0)},
		{userMsg.Enums["myenum"].Members["VAL_2"].Number, int32(1)},
		{userMsg.Enums["myenum"].Options["allow_alias"].Value, true},
		{userMsg.Oneofs["myoneof"].Name, "myoneof"},
		{userMsg.Fields["posts"].Repeated, true},
		{userMsg.Fields["fav_cat"].Optional, true},
		{userMsg.Fields["mymap"].Options["buf.validate.field.map.min_pairs"].Value, uint64(2)},
		{userMsg.Fields["mymap"].Options["buf.validate.field.map.max_pairs"].Value, uint64(4)},
		{userMsg.Fields["mymap"].Options["buf.validate.field.map.keys"].Value, "string : { min_len : 1 }"},
		{userMsg.Fields["mymap"].Options["buf.validate.field.map.values"].Value, "int64 : { gt : 1 in : [ 1 , 2 ] }"},
		{userMsg.Fields["reptest"].Options["buf.validate.field.repeated.min_items"].Value, uint64(1)},
		{userMsg.Fields["reptest"].Options["buf.validate.field.repeated.max_items"].Value, uint64(4)},
		{userMsg.Fields["reptest"].Options["buf.validate.field.repeated.items"].Value, "int32 : { gt : 1 in : [ 1 , 2 ] }"},
		{userMsg.Fields["timetest"].Options["buf.validate.field.timestamp.lt"].Value, fmt.Sprintf("seconds : %d nanos : 0", timePast.GetSeconds())},
		{userMsg.Fields["timetest2"].Options["buf.validate.field.timestamp.const"].Value, fmt.Sprintf("seconds : %d nanos : 0", timePast.GetSeconds())},
		// Non repeated options should be overridden
		{userMsg.Fields["fav_cat"].Options["myopt"].Value, false},
		// And separated from repeated options
		{len(userMsg.Fields["fav_cat"].Options), 1},
		{userMsg.Fields["fav_cat"].RepeatedOptions[0].Name, "buf.validate.field.cel"},
		{userMsg.Fields["fav_cat"].RepeatedOptions[0].Value, `id : "cel" message : "msg" expression : "expr"`},
		// Repeated options should be stacked
		{userMsg.Fields["fav_cat"].RepeatedOptions[2].Name, "repopt"},
		{userMsg.Fields["fav_cat"].RepeatedOptions[2].Value, true},
		{userMsg.Fields["fav_cat"].RepeatedOptions[3].Name, "repopt"},
		{userMsg.Fields["fav_cat"].RepeatedOptions[3].Value, true},
		{userMsg.Fields["fav_cat"].RepeatedOptions[4].Value, "tabby"},
		{userMsg.Fields["fav_cat"].RepeatedOptions[5].Value, "calico"},
		// Nested enums should have the correct name
		{out.Messages["GetUserRequest"].Fields["nestedenum"].TypeName, "User.nested.myenum"},
	}

	containsTests := []struct {
		Target   any
		Expected []any
	}{
		{out.Imports, []any{"buf/validate/validate.proto", "google/protobuf/empty.proto", "google/protobuf/timestamp.proto", "google/protobuf/field_mask.proto", "myapp/v1/post.proto"}},
		{out.Messages["UpdateUserRequest"].Fields, []any{"field_mask", "user"}},
		{out.Messages["User"].Fields, []any{"id", "created_at", "posts", "name"}},
		{out.Messages["GetUserRequest"].Fields, []any{"id"}},
		{out.Messages["GetUserResponse"].Fields, []any{"user"}},
		{out.Enums["myenum"].ReservedNames, []any{"name1", "name2"}},
		{out.Enums["myenum"].ReservedNumbers, []any{int32(10), int32(11)}},
		{out.Enums["myenum"].ReservedRanges, []any{sb.Range{20, 23}}},
		{out.Messages["User"].ReservedNames, []any{"name1", "name2"}},
		{out.Messages["User"].ReservedNumbers, []any{int32(20), int32(21)}},
		{out.Messages["User"].ReservedRanges, []any{sb.Range{22, 25}}},
	}

	shouldFail := []sb.FieldBuilder{
		sb.Repeated("nested_repeated_field", sb.Repeated("", sb.Timestamp(""))),
		sb.Repeated("repeated_map_field", sb.Map("", sb.String(""), sb.String(""))),
		sb.Repeated("non_scalar_unique", sb.Timestamp("")).Unique(),
		sb.Repeated("repeated_min_items>max_items", sb.Timestamp("")).MinItems(3).MaxItems(2),
		sb.Map("map_min_pairs>max_pairs", sb.String(""), sb.String("")).MinPairs(2).MaxPairs(1),
		sb.Map("map_repeated_as_value_type", sb.String(""), sb.Repeated("", sb.Timestamp(""))),
		sb.Map("map_as_map_value_type", sb.String(""), sb.Map("", sb.String(""), sb.String(""))),
		sb.Map("invalid_map_key", sb.Timestamp(""), sb.String("")),
		sb.String("string_in=notin").In("str").NotIn("str"),
		sb.String("string_min_len>max_len").MinLen(2).MaxLen(1),
		sb.String("string_max_len<min_len").MaxLen(1).MinLen(2),
		sb.String("string_max_len+len").MaxLen(1).Len(2),
		sb.String("string_min_len+len").MinLen(1).Len(2),
		sb.String("string_min_bytes>max_bytes").MinBytes(2).MaxBytes(1),
		sb.String("string_min_bytes<max_bytes").MaxBytes(1).MinBytes(2),
		sb.String("string_min_bytes+len_bytes").MinBytes(2).LenBytes(1),
		sb.String("string_multi_well_known").Ip().Ipv6(),
		sb.Int32("int32_lt=gt").Lt(1).Gt(1),
		sb.Int32("int32_lte<=gt").Lte(1).Gt(1),
		sb.Int32("int32_lt<=gte").Lt(1).Gte(1),
		sb.Int32("int32_lte<gte").Lte(0).Gte(1),
		sb.Int32("int32_gte>lte").Gte(2).Lte(1),
		sb.Int32("int32_gt>lte").Gt(2).Lte(1),
		sb.Int32("int32_gt>lt").Gt(2).Lt(1),
		sb.Int32("int32_non_finite").Finite(),
		sb.Int32("int32_lt+lte").Lt(2).Lte(2),
		sb.Int32("int32_gt+gte").Gt(2).Gte(2),
		sb.Timestamp("timestamp_lt+lt_now").Lt(&timePast).LtNow(),
		sb.Timestamp("timestamp_lte+lt_now").Lte(&timePast).LtNow(),
		sb.Timestamp("timestamp_gt_now+lt_now").GtNow().LtNow(),
		sb.Timestamp("timestamp_lte+lt").Lte(&timePast).Lt(&timePast),
		sb.Timestamp("timestamp_lt<=gt").Lt(&timePast).Gt(&timePast),
		sb.Timestamp("timestamp_lte<=gt").Lte(&timePast).Gt(&timePast),
		sb.Timestamp("timestamp_lt<=gte").Lt(&timePast).Gte(&timePast),
		sb.Timestamp("timestamp_lte<gte").Lte(&timePast).Gte(&timeFuture),
		sb.Timestamp("timestamp_lt<gt_now").Lt(&timePast).GtNow(),
		sb.Timestamp("timestamp_lte<gt_now").Lte(&timePast).GtNow(),
		sb.Timestamp("timestamp_gt>lt_now").Gt(&timeFuture).LtNow(),
		sb.Timestamp("timestamp_gte>lt_now").Gte(&timeFuture).LtNow(),
		sb.Timestamp("timestamp_gte>lte").Gte(&timeFuture).Lte(&timePast),
		sb.Timestamp("timestamp_gt>lte").Gt(&timeFuture).Lte(&timePast),
		sb.Timestamp("timestamp_gt>lt").Gt(&timeFuture).Lt(&timePast),
		sb.Duration("duration_lt+lte").Lt("1s").Lte("1s"),
		sb.Duration("duration_gt+gte").Gt("1s").Gte("1s"),
		sb.Duration("duration_lt<=gt").Lt("1s").Gt("1m"),
		sb.Duration("duration_lt<=gte").Lt("1s").Gte("1m"),
		sb.Duration("duration_lte<gte").Lte("1s").Gte("1m"),
		sb.Duration("duration_gte>lte").Gte("1m").Lte("1s"),
		sb.Duration("duration_gt>lte").Gt("1m").Lte("1s"),
		sb.Duration("duration_gt>lt").Gt("1m").Lt("1s"),
		sb.Duration("duration_in=notin").In("1s").NotIn("1s"),
		sb.String("const_with_extra_rules").Const("const").MinLen(2),
		sb.String("const_with_optional").Const("const").Optional(),
	}

	for _, test := range shouldFail {
		data, err := test.Build(1, sb.Set{})
		assert.Error(t, err, data.Name)
	}

	assert.NotContains(t, out.Messages, "Post")

	for _, test := range equalTests {
		assert.Equal(t, test.Expected, test.Target)
	}

	for _, test := range containsTests {
		for _, e := range test.Expected {
			assert.Contains(t, test.Target, e)
		}
	}
}

func ParseProtoFile(filePath string) FileData {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open generated proto file: %v", err)
	}
	defer file.Close()
	errRep := reporter.NewReporter(
		// Error handler
		func(err reporter.ErrorWithPos) error {
			return err
		},
		// Warning handler
		func(err reporter.ErrorWithPos) {
			fmt.Printf("[ WARN ]: %s", err.Error())
		})

	handler := reporter.NewHandler(errRep)
	ast, err := parser.Parse(filePath, file, handler)
	if err != nil {
		fmt.Print(err.Error())
	}

	data, err := parser.ResultFromAST(ast, false, handler)
	if err != nil {
		fmt.Print(err.Error())
	}

	desc := data.FileDescriptorProto()

	msgsMap := ExtractMessages(desc.GetMessageType())
	enumsMap := ExtractEnums(desc.GetEnumType())
	servicesMap := ExtractServices(desc.GetService())
	extensions := ExtractExtensions(desc.GetExtension())

	fileData := FileData{Package: desc.GetPackage(), Imports: desc.GetDependency(), Messages: msgsMap, Enums: enumsMap, Services: servicesMap, Extensions: extensions}

	return fileData
}

func ExtractOpts(opts []*descriptorpb.UninterpretedOption) (map[string]ProtoOption, []ProtoOption) {
	optionsMap := make(map[string]ProtoOption)
	areRepeated := make(map[string]struct{})
	flatOpts := []ProtoOption{}
	for _, o := range opts {
		name := ""
		var value any

		for i, part := range o.GetName() {
			if i != 0 {
				name += "."
			}
			name += part.GetNamePart()
		}

		if o.StringValue != nil {
			value = string(o.GetStringValue())
		} else if o.PositiveIntValue != nil {
			value = o.GetPositiveIntValue()
		} else if o.NegativeIntValue != nil {
			value = o.GetNegativeIntValue()
		} else if o.DoubleValue != nil {
			value = o.GetDoubleValue()
		} else if o.IdentifierValue != nil {
			strVal := o.GetIdentifierValue()

			strBool, err := strconv.ParseBool(strVal)
			if err != nil {
				value = strVal
			} else {
				value = strBool
			}

		} else if o.AggregateValue != nil {
			value = o.GetAggregateValue()
		}

		optsData := ProtoOption{
			Name: name, Value: value,
		}

		repeatedOpt, isRepeated := optionsMap[name]
		_, markedAsRepeated := areRepeated[name]

		if isRepeated && !markedAsRepeated {
			areRepeated[name] = struct{}{}
			markedAsRepeated = true
			flatOpts = append(flatOpts, repeatedOpt)
			delete(optionsMap, name)

		} else {
			optionsMap[name] = optsData
		}

		if markedAsRepeated {
			flatOpts = append(flatOpts, optsData)
		}

	}

	return optionsMap, flatOpts
}

func ExtractReservedNrs(ranges []*descriptorpb.DescriptorProto_ReservedRange) ([]int32, []sb.Range) {
	var reservedNumbers []int32
	var reservedRanges []sb.Range
	for _, r := range ranges {
		if r.Start != nil && r.End != nil {
			start := *r.Start
			end := *r.End

			if end-start == 1 {
				reservedNumbers = append(reservedNumbers, start)
			} else {
				// Somehow this is not consistent between enums and fields
				reservedRanges = append(reservedRanges, sb.Range{start, end - 1})
			}
		}
	}

	return reservedNumbers, reservedRanges
}

func ExtractEnumReservedNrs(ranges []*descriptorpb.EnumDescriptorProto_EnumReservedRange) ([]int32, []sb.Range) {
	var reservedNumbers []int32
	var reservedRanges []sb.Range
	for _, r := range ranges {
		if r.Start != nil && r.End != nil {
			start := *r.Start
			end := *r.End

			if end-start == 0 {
				reservedNumbers = append(reservedNumbers, start)
			} else {
				reservedRanges = append(reservedRanges, sb.Range{start, end})
			}
		}
	}

	return reservedNumbers, reservedRanges
}

func ExtractEnums(enums []*descriptorpb.EnumDescriptorProto) map[string]EnumData {
	data := make(map[string]EnumData)

	for _, e := range enums {
		var eData EnumData
		eData.Name = e.GetName()

		opts, repOpts := ExtractOpts(e.GetOptions().GetUninterpretedOption())

		eData.Options = opts
		eData.RepeatedOptions = repOpts
		eData.Members = make(map[string]EnumMember)

		for _, member := range e.GetValue() {
			var memData EnumMember
			memData.Number = member.GetNumber()
			eData.Members[member.GetName()] = memData
		}

		eData.ReservedNames = e.GetReservedName()

		resNrs, resRanges := ExtractEnumReservedNrs(e.GetReservedRange())
		eData.ReservedNumbers = resNrs
		eData.ReservedRanges = resRanges
		data[eData.Name] = eData
	}

	return data
}

func ExtractServices(services []*descriptorpb.ServiceDescriptorProto) map[string]ServiceData {
	data := make(map[string]ServiceData)

	for _, serv := range services {
		var service ServiceData

		service.Name = serv.GetName()
		service.Handlers = make(map[string]MethodData)
		opts, repOpts := ExtractOpts(serv.GetOptions().GetUninterpretedOption())

		service.Options = opts
		service.RepeatedOptions = repOpts
		for _, m := range serv.GetMethod() {
			var method MethodData

			method.Name = m.GetName()
			method.InputType = m.GetInputType()
			method.OutputType = m.GetOutputType()

			service.Handlers[method.Name] = method
		}

		data[service.Name] = service
	}

	return data
}

func ExtractFields(fields []*descriptorpb.FieldDescriptorProto) map[string]FieldData {
	fieldsMap := make(map[string]FieldData)
	for _, f := range fields {
		rawOpts := f.GetOptions().GetUninterpretedOption()

		opts, repeatedOpts := ExtractOpts(rawOpts)

		fieldData := FieldData{
			Number: f.GetNumber(), Optional: f.GetProto3Optional(), Repeated: f.GetLabel().String() == "LABEL_REPEATED",
			TypeName: f.GetTypeName(), Name: f.GetName(), Options: opts, RepeatedOptions: repeatedOpts,
		}

		fieldsMap[f.GetName()] = fieldData

	}

	return fieldsMap
}

func ExtractExtensions(exts []*descriptorpb.FieldDescriptorProto) map[string]ExtensionData {
	data := make(map[string]ExtensionData)

	for _, f := range exts {
		var fieldData FieldData

		extendee := f.GetExtendee()

		if _, exists := data[extendee]; !exists {
			data[extendee] = ExtensionData{Extendee: extendee, Fields: make(map[string]FieldData)}
		}

		opts, repOpts := ExtractOpts(f.GetOptions().GetUninterpretedOption())
		fieldData.Options = opts
		fieldData.RepeatedOptions = repOpts
		fieldData.TypeName = strings.ToLower(strings.TrimPrefix(f.GetType().String(), "TYPE_"))
		fieldData.Name = f.GetName()
		fieldData.Number = f.GetNumber()
		fieldData.Optional = f.GetProto3Optional()

		data[extendee].Fields[fieldData.Name] = fieldData
	}

	return data
}

func ExtractMessages(messages []*descriptorpb.DescriptorProto) map[string]MessageData {
	msgsMap := make(map[string]MessageData)
	for _, m := range messages {
		fieldsMap := ExtractFields(m.GetField())

		oneofs := ExtractOneofs(m.GetOneofDecl())
		enumData := ExtractEnums(m.GetEnumType())

		nestedMessages := ExtractMessages(m.GetNestedType())
		opts, repOpts := ExtractOpts(m.GetOptions().GetUninterpretedOption())

		resNrs, resRanges := ExtractReservedNrs(m.GetReservedRange())

		msgData := MessageData{
			Name: m.GetName(), Fields: fieldsMap, ReservedNames: m.GetReservedName(), Options: opts, RepeatedOptions: repOpts, ReservedNumbers: resNrs, ReservedRanges: resRanges, Enums: enumData, Oneofs: oneofs, Messages: nestedMessages,
		}

		msgsMap[m.GetName()] = msgData
	}

	return msgsMap
}

func ExtractOneofs(oneofs []*descriptorpb.OneofDescriptorProto) map[string]OneofData {
	data := make(map[string]OneofData)

	for _, o := range oneofs {
		opts, repOpts := ExtractOpts(o.GetOptions().GetUninterpretedOption())
		data[o.GetName()] = OneofData{Name: o.GetName(), Options: opts, RepeatedOptions: repOpts}
	}

	return data
}
