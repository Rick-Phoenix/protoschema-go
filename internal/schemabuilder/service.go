package schemabuilder

import (
	"errors"
	"fmt"
	"path"
	"strings"
)

type OptionExtensions struct {
	Service []CustomOption
	Message []CustomOption
	Field   []CustomOption
	File    []CustomOption
	OneOf   []CustomOption
}

type MessagesMap map[string]ProtoMessage

type ProtoService struct {
	Messages         []ProtoMessage
	Imports          Set
	ServiceOptions   []ProtoOption
	FileOptions      []ProtoOption
	OptionExtensions OptionExtensions
	Name             string
	FileOutput       string
	Handlers         []HandlerData
}

type HandlerData struct {
	Name     string
	Request  string
	Response string
}

type ProtoOption struct {
	Name  string
	Value any
}

type CustomOption struct {
	Name     string
	Type     string
	FieldNr  int
	Optional bool
	Repeated bool
}

type HandlersMap map[string]Handler

type Handler struct {
	Request  ProtoMessageSchema
	Response ProtoMessageSchema
}

type ProtoServiceSchema struct {
	Handlers         HandlersMap
	Messages         []ProtoMessageSchema
	ServiceOptions   []ProtoOption
	FileOptions      []ProtoOption
	OptionExtensions OptionExtensions
	Enums            ProtoEnumMap
}

var FileLocations = map[string]string{}

// Define handler messages (and even resource messages) in the handlers.
// Make a map to follow that, and for tohse that have not been listed as messages, add them.
// Otherwise, only use their name for the handler.
// Handlers should be a validated map regardless to avoid repetition.
// Maybe map of name like GetUser or GetUserRequest, mapped to a message schema (GetUserResponse)
func NewProtoService(resourceName string, s ProtoServiceSchema, basePath string) (ProtoService, error) {
	imports := make(Set)
	out := &ProtoService{FileOptions: s.FileOptions, ServiceOptions: s.ServiceOptions, Name: resourceName + "Service", Imports: imports, OptionExtensions: s.OptionExtensions}

	var messageErrors error

	for _, m := range s.Messages {
		message, err := NewProtoMessage(m, imports)
		out.Messages = append(out.Messages, message)

		if err != nil {
			messageErrors = errors.Join(messageErrors, IndentErrors(fmt.Sprintf("Errors for the %s message schema\n", resourceName), err))
		}
	}

	for name, h := range s.Handlers {
		out.Handlers = append(out.Handlers, HandlerData{Name: name, Request: h.Request.Name, Response: h.Response.Name})
	}

	if len(s.OptionExtensions.File)+len(s.OptionExtensions.Service)+len(s.OptionExtensions.Message)+len(s.OptionExtensions.Field)+len(s.OptionExtensions.OneOf) > 0 {
		imports["google/protobuf/descriptor.proto"] = present
	}

	fileOutput := path.Join(basePath, strings.ToLower(resourceName)+".proto")
	out.FileOutput = fileOutput
	FileLocations[resourceName] = fileOutput

	return *out, nil
}
