package schemabuilder

import (
	"errors"
	"fmt"
	"maps"
)

var present = struct{}{}

const indent = "  "
const indent2 = "    "

type Set map[string]struct{}

type Errors []error

type ProtoFieldData struct {
	Rules       map[string]any
	Options     []string
	ProtoType   string
	GoType      string
	Optional    bool
	FieldNr     uint
	Name        string
	Imports     []string
	Repeated    bool
	Required    bool
	IsNonScalar bool
}

type protoFieldInternal struct {
	rules           map[string]any
	options         map[string]any
	repeatedOptions []string
	optional        bool
	fieldNr         uint
	imports         []string
	protoType       string
	goType          string
	errors          error
	required        bool
	isNonScalar     bool
	ignore          string
	isEnum          bool
}

type ProtoFieldBuilder interface {
	Build(fieldName string, imports Set) (ProtoFieldData, error)
}

func (b *protoFieldInternal) Build(fieldName string, imports Set) (ProtoFieldData, error) {
	data := ProtoFieldData{Name: fieldName, ProtoType: b.protoType, GoType: b.goType, FieldNr: b.fieldNr, Rules: b.rules, IsNonScalar: b.isNonScalar, Optional: b.optional}

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

		protoName := b.protoType
		switch protoName {
		case "google.protobuf.Duration":
			protoName = "duration"
		case "google.protobuf.Timestamp":
			protoName = "timestamp"
		}

		if b.isEnum {
			protoName = "enum"
		}

		for rule, value := range b.rules {
			optsCollector[fmt.Sprintf("(buf.validate.field).%s.%s", protoName, rule)] = value
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
