package schemabuilder

import (
	"maps"
	"slices"
)

var present = struct{}{}

type Set map[string]struct{}

type ProtoService struct {
	Messages   []ProtoMessage
	FieldsFlat []string
	Imports    []string
	Options    map[string]string
}

type ProtoServiceSchema struct {
	Create, Get, Update, Delete *ServiceData
}

type ServiceData struct {
	Request  ProtoMessageSchema
	Response ProtoMessageSchema
}

func NewProtoService(resourceName string, s ProtoServiceSchema) ProtoService {
	out := &ProtoService{}
	imports := make(Set)
	if s.Get != nil {
		getRequest := NewProtoMessage("Get"+resourceName+"Request", s.Get.Request, imports)
		out.Messages = append(out.Messages, getRequest)
	}
	return *out
}

type ProtoFieldsMap map[string]ProtoFieldBuilder

type ProtoMessageSchema struct {
	Fields     ProtoFieldsMap
	Options    map[string]string
	CelOptions []CelFieldOpts
	Reserved   []int
	FieldMask  bool
}

type ProtoMessage struct {
	Name       string
	Fields     []ProtoFieldData
	Reserved   []int
	CelOptions []CelFieldOpts
	Options    map[string]string
}

func NewProtoMessage(messageName string, s ProtoMessageSchema, imports Set) ProtoMessage {
	var protoFields []ProtoFieldData
	for fieldName, fieldBuilder := range s.Fields {
		protoFields = append(protoFields, fieldBuilder.Build(fieldName, imports))
	}
	if s.FieldMask {
		imports["google/protobuf/field_mask.proto"] = struct{}{}
	}
	return ProtoMessage{Fields: protoFields, Name: messageName, Reserved: s.Reserved, Options: s.Options, CelOptions: s.CelOptions}
}

func ExtendProtoMessage(s ProtoMessageSchema, override ProtoMessageSchema) ProtoMessageSchema {
	newFields := make(ProtoFieldsMap)
	maps.Copy(newFields, s.Fields)
	maps.Copy(newFields, override.Fields)

	celOptions := slices.Concat(s.CelOptions, override.CelOptions)
	celOptions = DedupeNonComp(celOptions)

	reserved := slices.Concat(s.Reserved, override.Reserved)
	reserved = Dedupe(reserved)

	s.Fields = newFields
	s.Reserved = reserved

	return s
}

func OmitProtoMessage(s ProtoMessageSchema, keys []string) ProtoMessageSchema {
	newFields := make(ProtoFieldsMap)
	maps.Copy(newFields, s.Fields)

	for _, key := range keys {
		delete(newFields, key)
	}

	s.Fields = newFields

	return s
}

type ProtoField struct {
	Name       string
	Type       string
	Options    map[string]string
	CelOptions []CelFieldOpts
	Deprecated bool
}

type ProtoFieldData struct {
	Options    map[string]string
	CelOptions []CelFieldOpts
	ColType    string
	Nullable   bool
	FieldNr    int
	Name       string
	Imports    Set
	Deprecated bool
}

type protoFieldInternal struct {
	options    map[string]string
	celOptions []CelFieldOpts
	nullable   bool
	fieldNr    int
	imports    Set
	colType    string
	fieldMask  bool
	deprecated bool
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
	if b.fieldMask {
		imports["google/protobuf/field_mask.proto"] = present
	}
	maps.Copy(imports, b.imports)

	return ProtoFieldData{Name: fieldName, Options: b.options, ColType: "string", Nullable: b.nullable, FieldNr: b.fieldNr, CelOptions: b.celOptions}
}

func (b *ProtoFieldExternal) Nullable() *ProtoFieldExternal {
	b.nullable = true
	return b
}

func (b *ProtoFieldExternal) Deprecated() *ProtoFieldExternal {
	b.deprecated = true
	return b
}

func ProtoString(fieldNumber int) *ProtoFieldExternal {
	return &ProtoFieldExternal{&protoFieldInternal{fieldNr: fieldNumber, colType: "string"}}
}

func (b *ProtoFieldExternal) CelField(o CelFieldOpts) *ProtoFieldExternal {
	b.celOptions = append(b.celOptions, CelFieldOpts{
		Id: o.Id, Expression: o.Expression, Message: o.Message,
	})

	return b
}

// Make a helper that actually maps all these based on the col type for others
func (b *ProtoFieldExternal) Required() *ProtoFieldExternal {
	b.options["(buf.validate.field).required"] = "true"
	return b
}

func ProtoTimestamp(fieldNr int) *ProtoFieldExternal {
	return &ProtoFieldExternal{&protoFieldInternal{colType: "timestamp", fieldNr: fieldNr, fieldMask: true}}
}
