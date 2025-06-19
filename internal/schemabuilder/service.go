package schemabuilder

import (
	"errors"
	"fmt"
	"path"
	"slices"
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

// Message schema constructor for easy overriding
func NewProtoService(resourceName string, s ProtoServiceSchema, basePath string) (ProtoService, error) {
	imports := make(Set)
	var processedMessages []string

	messages := make([]ProtoMessageSchema, len(s.Messages))
	copy(messages, s.Messages)

	out := &ProtoService{FileOptions: s.FileOptions, ServiceOptions: s.ServiceOptions, Name: resourceName + "Service", Imports: imports, OptionExtensions: s.OptionExtensions}

	var messageErrors error

	for name, h := range s.Handlers {
		out.Handlers = append(out.Handlers, HandlerData{Name: name, Request: h.Request.Name, Response: h.Response.Name})
		if h.Request.ReferenceOnly {
			for _, im := range h.Request.Imports {
				imports[im] = present
			}
		} else if !slices.Contains(processedMessages, h.Request.Name) {
			messages = append(s.Messages, h.Request)
			processedMessages = append(processedMessages, h.Request.Name)
		}

		if h.Response.ReferenceOnly {
			for _, im := range h.Response.Imports {
				imports[im] = present
			}
		} else if !slices.Contains(processedMessages, h.Response.Name) {
			messages = append(s.Messages, h.Response)
			processedMessages = append(processedMessages, h.Response.Name)
		}

	}

	for _, m := range messages {
		message, err := NewProtoMessage(m, imports)
		out.Messages = append(out.Messages, message)
		processedMessages = append(processedMessages, m.Name)

		if err != nil {
			messageErrors = errors.Join(messageErrors, IndentErrors(fmt.Sprintf("Errors for the %s message schema\n", resourceName), err))
		}
	}

	if len(s.OptionExtensions.File)+len(s.OptionExtensions.Service)+len(s.OptionExtensions.Message)+len(s.OptionExtensions.Field)+len(s.OptionExtensions.OneOf) > 0 {
		imports["google/protobuf/descriptor.proto"] = present
	}

	fileOutput := path.Join(basePath, strings.ToLower(resourceName)+".proto")
	out.FileOutput = fileOutput
	FileLocations[resourceName] = fileOutput

	return *out, nil
}
