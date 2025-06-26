package schemabuilder

import (
	"errors"
	"fmt"
	"log"
	"maps"
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
	ResourceName     string
	Imports          Set
	OptionExtensions OptionExtensions
	Messages         []ProtoMessage
	Enums            []ProtoEnumGroup
	ServiceOptions   []ProtoOption
	FileOptions      []ProtoOption
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
	ResourceName     string
	Handlers         HandlersMap
	Messages         []ProtoMessageSchema
	ServiceOptions   []ProtoOption
	FileOptions      []ProtoOption
	OptionExtensions OptionExtensions
	Enums            []ProtoEnumGroup
}

func NewProtoService(s ProtoServiceSchema) (ProtoService, error) {
	imports := make(Set)
	processedMessages := make(Set)

	messages := make([]ProtoMessageSchema, len(s.Messages))
	copy(messages, s.Messages)

	out := &ProtoService{ResourceName: s.ResourceName, FileOptions: s.FileOptions, ServiceOptions: s.ServiceOptions, Imports: imports, OptionExtensions: s.OptionExtensions, Enums: s.Enums}

	var messageErrors error

	processMessage := func(m ProtoMessageSchema) {
		var errAgg error

		message, err := NewProtoMessage(m, imports)
		errAgg = errors.Join(errAgg, err)
		out.Messages = append(out.Messages, message)
		processedMessages[m.Name] = present

		if errAgg != nil {
			messageErrors = errors.Join(messageErrors, IndentErrors(fmt.Sprintf("Errors for the %s message schema", m.Name), errAgg))
		}
	}

	for _, m := range messages {
		processMessage(m)
	}

	handlerKeys := slices.SortedFunc(maps.Keys(s.Handlers), func(a, b string) int {
		methodsOrder := map[string]int{
			"Create": 0,
			"Get":    1,
			"Update": 2,
			"Delete": 3,
		}

		scoreA := 4
		scoreB := 4

		for method, value := range methodsOrder {
			if strings.HasPrefix(a, method) {
				scoreA = value
				break
			}

			if strings.HasPrefix(b, method) {
				scoreB = value
				break
			}
		}

		if scoreA == scoreB {
			return CompareString(a, b)
		}

		return scoreA - scoreB
	})

	for _, name := range handlerKeys {
		h := s.Handlers[name]

		out.Handlers = append(out.Handlers, HandlerData{Name: name, Request: h.Request.Name, Response: h.Response.Name})
		if _, seen := processedMessages[h.Request.Name]; !seen {
			if h.Request.ReferenceOnly {
				processedMessages[h.Request.Name] = present
				if h.Request.ImportPath != "" {
					imports[h.Request.ImportPath] = present
				}
			} else {
				processMessage(h.Request)
			}
		}

		if _, seen := processedMessages[h.Response.Name]; !seen {
			if h.Response.ReferenceOnly {
				processedMessages[h.Response.Name] = present
				if h.Response.ImportPath != "" {
					imports[h.Response.ImportPath] = present
				}
			} else {
				processMessage(h.Response)
			}
		}
	}

	if messageErrors != nil {
		log.Fatal(messageErrors)
	}

	if len(s.OptionExtensions.File)+len(s.OptionExtensions.Service)+len(s.OptionExtensions.Message)+len(s.OptionExtensions.Field)+len(s.OptionExtensions.OneOf) > 0 {
		imports["google/protobuf/descriptor.proto"] = present
	}

	return *out, nil
}
