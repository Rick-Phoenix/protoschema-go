package schemabuilder

// The protobuf extensions for a given file.
type Extensions struct {
	Service []ExtensionField
	Message []ExtensionField
	Field   []ExtensionField
	File    []ExtensionField
	OneOf   []ExtensionField
}

// A field belonging to a protobuf extension.
type ExtensionField struct {
	Name     string
	Type     string
	FieldNr  int
	Optional bool
	Repeated bool
}
