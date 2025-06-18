package schemabuilder

import (
	"errors"
	"fmt"
	"maps"
	"strings"
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
	Imports     Set
	Deprecated  bool
	Repeated    bool
	Required    bool
	IsNonScalar bool
}

type protoFieldInternal struct {
	options         map[string]string
	rules           map[string]any
	repeatedOptions []string
	optional        bool
	fieldNr         uint
	imports         Set
	protoType       string
	goType          string
	fieldMask       bool
	deprecated      bool
	errors          Errors
	required        bool
	isNonScalar     bool
}

type ProtoFieldBuilder interface {
	Build(fieldName string, imports Set) (ProtoFieldData, error)
}

func (b *protoFieldInternal) Build(fieldName string, imports Set) (ProtoFieldData, error) {
	if len(b.errors) > 0 {
		fieldErrors := strings.Builder{}
		for _, err := range b.errors {
			fieldErrors.WriteString(fmt.Sprintf("- %s\n", err.Error()))
		}

		return ProtoFieldData{}, errors.New(fieldErrors.String())
	}
	imports["buf/validate/validate.proto"] = present

	maps.Copy(imports, b.imports)

	options := GetOptions(b.options, b.repeatedOptions)

	return ProtoFieldData{Name: fieldName, Options: options, ProtoType: b.protoType, GoType: b.goType, Optional: b.optional, FieldNr: b.fieldNr, Rules: b.rules, IsNonScalar: b.isNonScalar}, nil
}

func (b *ProtoFieldExternal[BuilderT, ValueT]) Optional() *BuilderT {
	b.optional = true
	return b.self
}

type ProtoFieldExternal[BuilderT any, ValueT any] struct {
	*protoFieldInternal
	self *BuilderT
}

type GenericField[ValueT any] struct {
	*ProtoFieldExternal[GenericField[ValueT], ValueT]
}
