package protoschema

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

// The processed data from a FieldBuilder instance.
type FieldData struct {
	Rules         map[string]any
	Options       []string
	ProtoType     string
	ProtoBaseType string
	GoType        string
	Optional      bool
	FieldNr       uint32
	Name          string
	Imports       []string
	Repeated      bool
	IsMap         bool
	Required      bool
	IsNonScalar   bool
	MessageRef    *MessageSchema
	EnumRef       *EnumGroup
}

type protoFieldInternal struct {
	name            string
	rules           map[string]any
	options         map[string]any
	repeatedOptions []string
	optional        bool
	imports         []string
	protoType       string
	protoBaseType   string
	goType          string
	errors          error
	required        bool
	isNonScalar     bool
	repeated        bool
	isMap           bool
	isConst         bool
	messageRef      *MessageSchema
	enumRef         *EnumGroup
}

// The FieldBuilder interface, which is implemented by the various field constructors.
type FieldBuilder interface {
	Build(fieldNr uint32, imports Set) (FieldData, error)
	GetData() FieldData
	IsMap() bool
	IsRepeated() bool
	IsNonScalar() bool
	GetGoType() string
	GetName() string
	GetMessageRef() *MessageSchema
}

func (b *protoFieldInternal) IsNonScalar() bool {
	return b.isNonScalar
}

func (b *protoFieldInternal) GetMessageRef() *MessageSchema {
	return b.messageRef
}

func (b *protoFieldInternal) IsMap() bool {
	return b.isMap
}

func (b *protoFieldInternal) IsRepeated() bool {
	return b.repeated
}

func (b *protoFieldInternal) GetGoType() string {
	return b.goType
}

func (b *protoFieldInternal) GetName() string {
	return b.name
}

func (b *protoFieldInternal) GetData() FieldData {
	return FieldData{
		Name: b.name, ProtoType: b.protoType, ProtoBaseType: b.protoBaseType, Rules: maps.Clone(b.rules),
		Imports:  slices.Clone(b.imports),
		Repeated: b.repeated, Required: b.required, IsNonScalar: b.isNonScalar, Optional: b.optional,
		GoType: b.goType, IsMap: b.isMap, MessageRef: b.messageRef,
	}
}

func (b *protoFieldInternal) Build(fieldNr uint32, imports Set) (FieldData, error) {
	data := FieldData{
		Name: b.name, ProtoType: b.protoType, GoType: b.goType, FieldNr: fieldNr,
		Rules: b.rules, IsNonScalar: b.isNonScalar, Optional: b.optional, ProtoBaseType: b.protoBaseType, IsMap: b.isMap,
		MessageRef: b.messageRef,
	}

	if data.ProtoBaseType == "" {
		data.ProtoBaseType = data.ProtoType
	}

	var errAgg error
	if b.errors != nil {
		errAgg = errors.Join(errAgg, b.errors)
	}

	for _, v := range b.imports {
		imports[v] = present
	}

	options := make([]string, len(b.repeatedOptions))
	copy(options, b.repeatedOptions)

	optsCollector := make(map[string]any)
	maps.Copy(optsCollector, b.options)

	if b.isConst {
		if len(b.rules) > 1 {
			errAgg = errors.Join(errAgg, fmt.Errorf("A constant field cannot have extra rules."))
		}
		if b.optional {
			errAgg = errors.Join(errAgg, fmt.Errorf("A constant field cannot be optional."))
		}
	}

	if len(b.rules) > 0 {
		imports["buf/validate/validate.proto"] = present

		for rule, value := range b.rules {
			optsCollector[fmt.Sprintf("(buf.validate.field).%s.%s", data.ProtoBaseType, rule)] = value
		}
	}

	options, err := getOptions(optsCollector, options)
	if err != nil {
		errAgg = errors.Join(errAgg, err)
	}

	data.Options = options

	if errAgg != nil {
		return data, errAgg
	}

	return data, nil
}

// The generic ProtoField type, which contains the methods shared among all FieldBuilder implementations.
type ProtoField[BuilderT any] struct {
	*protoFieldInternal
	self *BuilderT
}

// An instance of a generic protobuf field. Mostly used for imported message fields.
type GenericField struct {
	*ProtoField[GenericField]
}
