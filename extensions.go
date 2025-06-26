package schemabuilder

type Extensions struct {
	Service []ExtensionField
	Message []ExtensionField
	Field   []ExtensionField
	File    []ExtensionField
	OneOf   []ExtensionField
}

type ExtensionField struct {
	Name     string
	Type     string
	FieldNr  int
	Optional bool
	Repeated bool
}
