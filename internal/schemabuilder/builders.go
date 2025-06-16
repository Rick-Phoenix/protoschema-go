package schemabuilder

import (
	"fmt"
	"log"
	"maps"
	"path"
	"reflect"
	"slices"
)

var present = struct{}{}

type Set map[string]struct{}

type MessagesMap map[string]ProtoMessage

type ProtoService struct {
	Messages   MessagesMap
	Imports    Set
	Options    map[string]string
	Name       string
	FileOutput string
	Handlers   []string
}

type ProtoServiceSchema struct {
	Create, Get, Update, Delete *ServiceData
	Resource                    ProtoMessageSchema
	Options                     map[string]string
}

type ServiceData struct {
	Request  ProtoMessageSchema
	Response ProtoMessageSchema
}

var FileLocations = map[string]string{}

func NewProtoService(resourceName string, s ProtoServiceSchema, basePath string) ProtoService {
	imports := make(Set)
	messages := make(MessagesMap)
	out := &ProtoService{Options: s.Options, Name: resourceName + "Service", Imports: imports, Messages: messages}

	messages[resourceName] = NewProtoMessage(s.Resource, imports)

	if s.Get != nil {
		out.Handlers = append(out.Handlers, "Get"+resourceName)
		requestName := "Get" + resourceName + "Request"
		getRequest := NewProtoMessage(s.Get.Request, imports)
		responseName := "Get" + resourceName + "Response"
		getResponse := NewProtoMessage(s.Get.Response, imports)
		messages[requestName] = getRequest
		messages[responseName] = getResponse
	}

	fileOutput := path.Join(basePath, resourceName+".proto")
	out.FileOutput = fileOutput
	FileLocations[resourceName] = fileOutput

	return *out
}

type ProtoFieldsMap map[string]ProtoFieldBuilder

type ProtoMessageSchema struct {
	Fields     ProtoFieldsMap
	Options    map[string]string
	CelOptions []CelFieldOpts
	Reserved   []int
}

type ProtoMessage struct {
	Fields     []ProtoFieldData
	Reserved   []int
	CelOptions []CelFieldOpts
	Options    map[string]string
}

func NewProtoMessage(s ProtoMessageSchema, imports Set) ProtoMessage {
	var protoFields []ProtoFieldData
	for fieldName, fieldBuilder := range s.Fields {
		protoFields = append(protoFields, fieldBuilder.Build(fieldName, imports))
	}

	return ProtoMessage{Fields: protoFields, Reserved: s.Reserved, Options: s.Options, CelOptions: s.CelOptions}
}

func ExtendProtoMessage(s ProtoMessageSchema, override *ProtoMessageSchema) *ProtoMessageSchema {
	if override == nil {
		return &s
	}
	newFields := make(ProtoFieldsMap)
	maps.Copy(newFields, s.Fields)
	maps.Copy(newFields, override.Fields)

	newOptions := make(map[string]string)
	maps.Copy(newOptions, s.Options)
	maps.Copy(newOptions, override.Options)

	newCelOptions := slices.Concat(s.CelOptions, override.CelOptions)
	newCelOptions = DedupeNonComp(newCelOptions)

	reserved := slices.Concat(s.Reserved, override.Reserved)
	reserved = Dedupe(reserved)

	s.Fields = newFields
	s.Reserved = reserved
	s.Options = newOptions
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
}

type ProtoFieldBuilder interface {
	Build(fieldName string, imports Set) ProtoFieldData
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

func GetCelOptions(opts []CelFieldOpts) []string {
	flatOpts := []string{}

	for i, opt := range opts {
		stringOpt := fmt.Sprintf(
			`(buf.validate.field).cel = {
			id: %q
			message: %q
			expression: %q
		}`,
			opt.Id, opt.Message, opt.Expression)
		if i < len(opts)-1 {
			stringOpt += ", "
		}

		flatOpts = append(flatOpts, stringOpt)
	}

	return flatOpts
}

func (b *protoFieldInternal) Build(fieldName string, imports Set) ProtoFieldData {

	imports["buf/validate/validate.proto"] = present

	if b.repeated {
		if b.optional {
			log.Fatalf("Field %s cannot be repeated and optional.", fieldName)
		}
	}

	maps.Copy(imports, b.imports)

	options := GetOptions(b.options, b.celOptions)

	return ProtoFieldData{Name: fieldName, Options: options, FieldType: b.fieldType, Optional: b.optional, FieldNr: b.fieldNr, Repeated: b.repeated}
}

func (b *ProtoFieldExternal) Optional() *ProtoFieldExternal {

	b.optional = true
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
	b.repeated = true
	return b
}

func ProtoString(fieldNumber int) *ProtoFieldExternal {
	imports := make(Set)
	options := make(map[string]string)
	return &ProtoFieldExternal{&protoFieldInternal{fieldNr: fieldNumber, fieldType: "string", imports: imports, options: options}}
}

// To refine with proto type and go type
func (b *ProtoFieldExternal) Const(val any) *ProtoFieldExternal {
	valType := reflect.TypeOf(val).String()
	if valType != b.fieldType {
		log.Fatalf("The type for const does not match.")
	}
	return b
}

// Make a helper that actually maps all these based on the col type for others
// So that for example MaxLen accepts an int, and then the map's values should be strings or any
func (b *ProtoFieldExternal) Required() *ProtoFieldExternal {
	b.optional = false
	b.options["(buf.validate.field).required"] = "true"
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
