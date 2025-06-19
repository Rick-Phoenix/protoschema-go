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

// Allow for protoempty as a response
type ProtoService struct {
	Messages         MessagesMap
	Imports          Set
	ServiceOptions   []ProtoOption
	FileOptions      []ProtoOption
	OptionExtensions OptionExtensions
	Name             string
	FileOutput       string
	Handlers         []string
}

type ProtoOption struct {
	Name  string
	Value string
}

type CustomOption struct {
	Name     string
	Type     string
	FieldNr  int
	Optional bool
	Repeated bool
}

type ProtoServiceSchema struct {
	Create, Get, Update, Delete *ServiceData
	Resource                    ProtoMessageSchema
	ServiceOptions              []ProtoOption
	FileOptions                 []ProtoOption
	OptionExtensions            OptionExtensions
	Enums                       ProtoEnumMap
}

type ServiceData struct {
	Request  ProtoMessageSchema
	Response ProtoMessageSchema
}

var FileLocations = map[string]string{}

func NewProtoService(resourceName string, s ProtoServiceSchema, basePath string) (ProtoService, error) {
	imports := make(Set)
	messages := make(MessagesMap)
	out := &ProtoService{FileOptions: s.FileOptions, ServiceOptions: s.ServiceOptions, Name: resourceName + "Service", Imports: imports, Messages: messages, OptionExtensions: s.OptionExtensions}

	var messageErrors error
	message, err := NewProtoMessage(s.Resource, imports)
	messages[resourceName] = message

	if err != nil {
		messageErrors = errors.Join(messageErrors, IndentErrors(fmt.Sprintf("Errors for the %s message schema\n", resourceName), err))
	}

	if len(s.OptionExtensions.File)+len(s.OptionExtensions.Service)+len(s.OptionExtensions.Message)+len(s.OptionExtensions.Field)+len(s.OptionExtensions.OneOf) > 0 {
		imports["google/protobuf/descriptor.proto"] = present
	}

	// if s.Get != nil {
	// 	out.Handlers = append(out.Handlers, "Get"+resourceName)
	// 	requestName := "Get" + resourceName + "Request"
	// 	// getRequest := NewProtoMessage(s.Get.Request, imports)
	// 	responseName := "Get" + resourceName + "Response"
	// 	// getResponse := NewProtoMessage(s.Get.Response, imports)
	// 	messages[requestName] = getRequest
	// 	messages[responseName] = getResponse
	// }

	fileOutput := path.Join(basePath, strings.ToLower(resourceName)+".proto")
	out.FileOutput = fileOutput
	FileLocations[resourceName] = fileOutput

	return *out, nil
}
