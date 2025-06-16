package schemabuilder

import (
	"maps"
	"path"
	"slices"
)

var present = struct{}{}

type Set map[string]struct{}

type MessagesMap map[string]ProtoMessage

type ProtoService struct {
	Messages MessagesMap
	Imports  Set
	Options  map[string]string
	Name     string
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
		messageName := "Get" + resourceName + "Request"
		getRequest := NewProtoMessage(s.Get.Request, imports)
		messages[messageName] = getRequest
	}

	FileLocations[resourceName] = path.Join(basePath, resourceName+"_service.proto")
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
	Options    map[string]string
	CelOptions []CelFieldOpts
	FieldType  string
	Nullable   bool
	FieldNr    int
	Name       string
	Imports    Set
	Deprecated bool
	Repeated   bool
}

type protoFieldInternal struct {
	options    map[string]string
	celOptions []CelFieldOpts
	nullable   bool
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

func (b *protoFieldInternal) Build(fieldName string, imports Set) ProtoFieldData {
	switch b.fieldType {
	case "fieldmask":
		imports["google/protobuf/field_mask.proto"] = present
	case "timestamp":
		imports["google/protobuf/timestamp.proto"] = present
	}

	if b.repeated {
		b.nullable = false
	}

	maps.Copy(imports, b.imports)

	return ProtoFieldData{Name: fieldName, Options: b.options, FieldType: "string", Nullable: b.nullable, FieldNr: b.fieldNr, CelOptions: b.celOptions, Repeated: b.repeated}
}

func (b *ProtoFieldExternal) Nullable() *ProtoFieldExternal {
	b.nullable = true
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

// Make a helper that actually maps all these based on the col type for others
// So that for example MaxLen accepts an int, and then the map's values should be strings or any
func (b *ProtoFieldExternal) Required() *ProtoFieldExternal {
	b.options["(buf.validate.field).required"] = "true"
	return b
}

func ProtoTimestamp(fieldNr int) *ProtoFieldExternal {
	imports := make(Set)
	options := make(map[string]string)
	return &ProtoFieldExternal{&protoFieldInternal{fieldNr: fieldNr, fieldType: "timestamp", imports: imports, options: options}}
}

func FieldMask(fieldNr int) *ProtoFieldExternal {
	imports := make(Set)
	options := make(map[string]string)
	return &ProtoFieldExternal{&protoFieldInternal{fieldNr: fieldNr, fieldType: "fieldmask", imports: imports, options: options}}
}

// Way to get the import paths
func ExternalType(fieldNr int, name string) *ProtoFieldExternal {
	imports := make(Set)
	options := make(map[string]string)
	importPath := FileLocations[name]
	imports[importPath] = present
	return &ProtoFieldExternal{&protoFieldInternal{fieldNr: fieldNr, fieldType: "external", imports: imports, options: options}}
}
