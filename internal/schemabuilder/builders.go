package schemabuilder

import (
	"errors"
	"fmt"
	"maps"
	"path"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

var present = struct{}{}

type Set map[string]struct{}

type MessagesMap map[string]ProtoMessage

type OptionExtensions struct {
	Service []CustomOption
	Message []CustomOption
	Field   []CustomOption
	File    []CustomOption
}

type ProtoService struct {
	Messages         MessagesMap
	Imports          Set
	ServiceOptions   []ProtoOption
	FileOptions      []ProtoOption
	OptionExtensions OptionExtensions
	Name             string
	FileOutput       string
	Handlers         []string
}

type ProtoOption struct {
	Name  string
	Value string
}

type CustomOption struct {
	Name     string
	Type     string
	FieldNr  int
	Optional bool
	Repeated bool
}

type ProtoServiceSchema struct {
	Create, Get, Update, Delete *ServiceData
	Resource                    ProtoMessageSchema
	ServiceOptions              []ProtoOption
	FileOptions                 []ProtoOption
	OptionExtensions            OptionExtensions
}

type ServiceData struct {
	Request  ProtoMessageSchema
	Response ProtoMessageSchema
}

var FileLocations = map[string]string{}

func NewProtoService(resourceName string, s ProtoServiceSchema, basePath string) (ProtoService, error) {
	imports := make(Set)
	messages := make(MessagesMap)
	out := &ProtoService{FileOptions: s.FileOptions, ServiceOptions: s.ServiceOptions, Name: resourceName + "Service", Imports: imports, Messages: messages, OptionExtensions: s.OptionExtensions}

	message, err := NewProtoMessage(s.Resource, imports)
	messages[resourceName] = message

	if len(err) > 0 {
		messageErrors := strings.Builder{}
		messageErrors.WriteString(fmt.Sprintf("The following errors occurred for the %s message schema:\n", resourceName))
		for _, errGroup := range err {
			messageErrors.WriteString(IndentString(errGroup.Error()))
		}

		return ProtoService{}, errors.New(messageErrors.String())
	}

	if len(s.OptionExtensions.File)+len(s.OptionExtensions.Service)+len(s.OptionExtensions.Message)+len(s.OptionExtensions.Field) > 0 {
		imports["google/protobuf/descriptor.proto"] = present
	}

	// if s.Get != nil {
	// 	out.Handlers = append(out.Handlers, "Get"+resourceName)
	// 	requestName := "Get" + resourceName + "Request"
	// 	// getRequest := NewProtoMessage(s.Get.Request, imports)
	// 	responseName := "Get" + resourceName + "Response"
	// 	// getResponse := NewProtoMessage(s.Get.Response, imports)
	// 	messages[requestName] = getRequest
	// 	messages[responseName] = getResponse
	// }

	fileOutput := path.Join(basePath, resourceName+".proto")
	out.FileOutput = fileOutput
	FileLocations[resourceName] = fileOutput

	return *out, nil
}

type ProtoFieldsMap map[string]ProtoFieldBuilder

type ProtoMessageSchema struct {
	Fields     ProtoFieldsMap
	Options    []MessageOption
	CelOptions []CelFieldOpts
	Reserved   []int
}

type ProtoMessage struct {
	Fields     []ProtoFieldData
	Reserved   []int
	CelOptions []CelFieldOpts
	Options    []MessageOption
}

func NewProtoMessage(s ProtoMessageSchema, imports Set) (ProtoMessage, Errors) {
	var protoFields []ProtoFieldData
	var errors Errors

	for fieldName, fieldBuilder := range s.Fields {
		field, err := fieldBuilder.Build(fieldName, imports)
		if err != nil {
			errors = append(errors, err)
		} else {
			protoFields = append(protoFields, field)
		}
	}

	if len(errors) > 0 {
		return ProtoMessage{}, errors
	}

	return ProtoMessage{Fields: protoFields, Reserved: s.Reserved, Options: s.Options, CelOptions: s.CelOptions}, nil
}

func ExtendProtoMessage(s ProtoMessageSchema, override *ProtoMessageSchema) *ProtoMessageSchema {
	if override == nil {
		return &s
	}
	newFields := make(ProtoFieldsMap)
	maps.Copy(newFields, s.Fields)
	maps.Copy(newFields, override.Fields)

	newCelOptions := slices.Concat(s.CelOptions, override.CelOptions)
	newCelOptions = DedupeNonComp(newCelOptions)

	reserved := slices.Concat(s.Reserved, override.Reserved)
	reserved = Dedupe(reserved)

	s.Fields = newFields
	s.Reserved = reserved
	s.Options = override.Options
	s.CelOptions = newCelOptions

	return &s
}

func OmitProtoMessage(s ProtoMessageSchema, keys []string) *ProtoMessageSchema {
	newFields := make(ProtoFieldsMap)
	maps.Copy(newFields, s.Fields)

	for _, key := range keys {
		delete(newFields, key)
	}

	s.Fields = newFields

	return &s
}

type ProtoFieldData struct {
	Options    []string
	FieldType  string
	Optional   bool
	FieldNr    int
	Name       string
	Imports    Set
	Deprecated bool
	Repeated   bool
}

type protoFieldInternal struct {
	options    map[string]string
	celOptions []CelFieldOpts
	optional   bool
	fieldNr    int
	imports    Set
	fieldType  string
	fieldMask  bool
	deprecated bool
	repeated   bool
	errors     Errors
	required   bool
}

type ProtoFieldBuilder interface {
	Build(fieldName string, imports Set) (ProtoFieldData, error)
}

type ProtoFieldExternal struct {
	*protoFieldInternal
}

type CelFieldOpts struct {
	Id, Message, Expression string
}

func GetOptions(optsMap map[string]string, celOpts []CelFieldOpts) []string {
	flatOpts := []string{}
	optNames := slices.Collect(maps.Keys(optsMap))

	for i, name := range optNames {
		stringOpt := name + " = " + optsMap[name]
		if i < len(optNames)-1 || len(celOpts) > 0 {
			stringOpt += ", "
		}

		flatOpts = append(flatOpts, stringOpt)
	}

	flatCelOpts := GetCelOptions(celOpts)

	flatOpts = slices.Concat(flatOpts, flatCelOpts)

	return flatOpts
}

type MessageOption struct {
	Name  string
	Value string
}

type Errors []error

func MessageCelOption(o CelFieldOpts) MessageOption {
	return MessageOption{
		Name: "(buf.validate.field).cel", Value: GetCelOption(o),
	}
}

var DisableValidation = MessageOption{
	Name: "(buf.validate.message).disabled", Value: "true",
}

func GetCelOption(opt CelFieldOpts) string {
	return fmt.Sprintf(
		`{
			id: %q
			message: %q
			expression: %q
		}`,
		opt.Id, opt.Message, opt.Expression)

}

func GetCelOptions(opts []CelFieldOpts) []string {
	flatOpts := []string{}

	for i, opt := range opts {
		stringOpt := fmt.Sprintf(
			`(buf.validate.field).cel = %s`,
			GetCelOption(opt))
		if i < len(opts)-1 {
			stringOpt += ", "
		}

		flatOpts = append(flatOpts, stringOpt)
	}

	return flatOpts
}

const indent = "  "
const indent2 = "    "
const indent3 = "      "

func (b *protoFieldInternal) Build(fieldName string, imports Set) (ProtoFieldData, error) {
	if len(b.errors) > 0 {
		fieldErrors := strings.Builder{}
		fieldErrors.WriteString(fmt.Sprintf("Errors for field %s:\n", fieldName))
		for _, err := range b.errors {
			fieldErrors.WriteString(fmt.Sprintf("%s- %s\n", indent, err.Error()))
		}

		return ProtoFieldData{}, errors.New(fieldErrors.String())
	}
	imports["buf/validate/validate.proto"] = present

	maps.Copy(imports, b.imports)

	options := GetOptions(b.options, b.celOptions)

	return ProtoFieldData{Name: fieldName, Options: options, FieldType: b.fieldType, Optional: b.optional, FieldNr: b.fieldNr, Repeated: b.repeated}, nil
}

func (b *ProtoFieldExternal) Optional() *ProtoFieldExternal {
	if b.repeated {
		b.errors = append(b.errors, fmt.Errorf("A field cannot be repeated and optional."))
	}
	if b.required {
		b.errors = append(b.errors, fmt.Errorf("A field cannot be required and optional."))
	}
	b.optional = true
	return b
}

func (b *ProtoFieldExternal) IgnoreIfUnpopulated() *ProtoFieldExternal {
	b.options["(buf.validate.field).ignore"] = "IGNORE_IF_UNPOPULATED"
	return b
}

func (b *ProtoFieldExternal) IgnoreIfDefaultValue() *ProtoFieldExternal {
	b.options["(buf.validate.field).ignore"] = "IGNORE_IF_DEFAULT_VALUE"
	return b
}

func (b *ProtoFieldExternal) IgnoreAlways() *ProtoFieldExternal {
	b.options["(buf.validate.field).ignore"] = "IGNORE_ALWAYS"
	return b
}

func (b *ProtoFieldExternal) Deprecated() *ProtoFieldExternal {
	b.options["deprecated"] = "true"
	return b
}

func (b *ProtoFieldExternal) CelField(o CelFieldOpts) *ProtoFieldExternal {
	b.celOptions = append(b.celOptions, CelFieldOpts{
		Id: o.Id, Expression: o.Expression, Message: o.Message,
	})

	return b
}

func (b *ProtoFieldExternal) Repeated() *ProtoFieldExternal {
	if b.optional {
		b.errors = append(b.errors, fmt.Errorf("A field cannot be repeated and optional."))
	}
	b.repeated = true
	return b
}

// To refine with proto type and go type
func (b *ProtoFieldExternal) Const(val any) *ProtoFieldExternal {
	valType := reflect.TypeOf(val).String()
	if valType != b.fieldType {
		err := fmt.Errorf("The type for const does not match.\nField type: %s\nConst type: %s", b.fieldType, valType)
		b.errors = append(b.errors, err)
	}
	return b
}

func (b *ProtoFieldExternal) Required() *ProtoFieldExternal {
	if b.optional {
		b.errors = append(b.errors, fmt.Errorf("A field cannot be required and optional."))
	}
	b.options["(buf.validate.field).required"] = "true"
	b.required = true
	return b
}

func (b *ProtoFieldExternal) Example(e string) *ProtoFieldExternal {
	// Make this specific to the single validators
	b.options["(buf.validate.field).timestamp.example"] = e
	return b
}

func ProtoTimestamp(fieldNr int) *ProtoFieldExternal {
	imports := make(Set)
	options := make(map[string]string)
	imports["google/protobuf/timestamp.proto"] = present
	return &ProtoFieldExternal{&protoFieldInternal{fieldNr: fieldNr, fieldType: "google.protobuf.Timestamp", imports: imports, options: options}}
}

func FieldMask(fieldNr int) *ProtoFieldExternal {
	imports := make(Set)
	options := make(map[string]string)
	imports["google/protobuf/field_mask.proto"] = present
	return &ProtoFieldExternal{&protoFieldInternal{fieldNr: fieldNr, fieldType: "fieldmask", imports: imports, options: options}}
}

func ImportedType(fieldNr int, name string, importPath string) *ProtoFieldExternal {
	imports := make(Set)
	options := make(map[string]string)
	imports[importPath] = present
	return &ProtoFieldExternal{&protoFieldInternal{fieldNr: fieldNr, fieldType: name, imports: imports, options: options}}
}

func InternalType(fieldNr int, name string) *ProtoFieldExternal {
	imports := make(Set)
	options := make(map[string]string)
	return &ProtoFieldExternal{&protoFieldInternal{fieldNr: fieldNr, fieldType: name, imports: imports, options: options}}
}

type FieldWithLength struct{}

type StringField struct {
	*ProtoFieldExternal
}

func (b *StringField) MinLen(n int) *StringField {
	b.options["(buf.validate.field).string.min_len"] = strconv.Itoa(n)
	return b
}

func ProtoString(fieldNumber int) *StringField {
	imports := make(Set)
	options := make(map[string]string)
	return &StringField{&ProtoFieldExternal{&protoFieldInternal{fieldNr: fieldNumber, fieldType: "string", imports: imports, options: options}}}
}
