package schemabuilder

type ProtoOneOfData struct {
	Name    string
	Choices []ProtoFieldData
	Options []ProtoOption
}

type ProtoOneOfSchema struct {
	Name    string
	Choices ProtoOneOfsMap
	Options []ProtoOption
}

type ProtoOneOfsMap map[string]ProtoFieldBuilder

// This applies to any of the items
var OneOfRequired = ProtoOption{
	Name:  "(buf.validate.oneof).required",
	Value: "true",
}
