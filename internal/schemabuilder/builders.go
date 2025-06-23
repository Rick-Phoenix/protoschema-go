package schemabuilder

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

var present = struct{}{}

const (
	indent  = "  "
	indent2 = "    "
)

type Set map[string]struct{}

type Errors []error

type ProtoFieldData struct {
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
	Required      bool
	IsNonScalar   bool
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
}

type ProtoFieldBuilder interface {
	Build(fieldNr uint32, imports Set) (ProtoFieldData, error)
	GetData() ProtoFieldData
}

func (b *protoFieldInternal) GetData() ProtoFieldData {
	return ProtoFieldData{
		Name: b.name, ProtoType: b.protoType, ProtoBaseType: b.protoBaseType, Rules: maps.Clone(b.rules),
		Imports:  slices.Clone(b.imports),
		Repeated: b.repeated, Required: b.required, IsNonScalar: b.isNonScalar, Optional: b.optional,
	}
}

func (b *protoFieldInternal) Build(fieldNr uint32, imports Set) (ProtoFieldData, error) {
	data := ProtoFieldData{
		Name: b.name, ProtoType: b.protoType, GoType: b.goType, FieldNr: fieldNr,
		Rules: b.rules, IsNonScalar: b.isNonScalar, Optional: b.optional, ProtoBaseType: b.protoBaseType,
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

	var options []string

	optsCollector := make(map[string]any)
	maps.Copy(optsCollector, b.options)

	if len(b.rules) > 0 {
		imports["buf/validate/validate.proto"] = present

		for rule, value := range b.rules {
			optsCollector[fmt.Sprintf("(buf.validate.field).%s.%s", data.ProtoBaseType, rule)] = value
		}
	}

	options, err := GetOptions(optsCollector, b.repeatedOptions)
	if err != nil {
		errAgg = errors.Join(errAgg, err)
	}

	data.Options = options

	if errAgg != nil {
		return data, errAgg
	}

	return data, nil
}

type ProtoFieldExternal[BuilderT any, ValueT any] struct {
	*protoFieldInternal
	self *BuilderT
}

type GenericField[ValueT any] struct {
	*ProtoFieldExternal[GenericField[ValueT], ValueT]
}
