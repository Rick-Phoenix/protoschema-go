package schemabuilder

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
	imports         []string
	protoType       string
	goType          string
	deprecated      bool
	errors          error
	required        bool
	isNonScalar     bool
}

type ProtoFieldBuilder interface {
	Build(fieldName string, imports Set) (ProtoFieldData, error)
}

func (b *protoFieldInternal) Build(fieldName string, imports Set) (ProtoFieldData, error) {
	if b.errors != nil {
		return ProtoFieldData{}, b.errors
	}

	for _, v := range b.imports {
		imports[v] = present
	}

	options := GetOptions(b.options, b.repeatedOptions)

	return ProtoFieldData{Name: fieldName, Options: options, ProtoType: b.protoType, GoType: b.goType, FieldNr: b.fieldNr, Rules: b.rules, IsNonScalar: b.isNonScalar, Optional: b.optional}, nil
}

type ProtoFieldExternal[BuilderT any, ValueT any] struct {
	*protoFieldInternal
	self *BuilderT
}

type GenericField[ValueT any] struct {
	*ProtoFieldExternal[GenericField[ValueT], ValueT]
}
