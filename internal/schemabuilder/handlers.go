package schemabuilder

type HandlersMap map[string]ProtoMessageSchema

type Handler struct {
	Request  string
	Response string
}

var ValidMethods = []string{"Get", "Create", "Update", "Delete"}
