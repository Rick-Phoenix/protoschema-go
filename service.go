package schemabuilder

import (
	"errors"
	"fmt"
	"log"
	"maps"
	"slices"
	"strings"
)

type HandlerData struct {
	Name     string
	Request  string
	Response string
}

type HandlersMap map[string]Handler

type Handler struct {
	Request  MessageSchema
	Response MessageSchema
}

type ServiceData struct {
	ResourceName   string
	Imports        Set
	Extensions     Extensions
	Messages       []MessageData
	Enums          []ProtoEnumGroup
	ServiceOptions []ProtoOption
	FileOptions    []ProtoOption
	Handlers       []HandlerData
}

type ServiceSchema struct {
	Resource         MessageSchema
	Handlers         HandlersMap
	Messages         []MessageSchema
	ServiceOptions   []ProtoOption
	FileOptions      []ProtoOption
	OptionExtensions Extensions
	Enums            []ProtoEnumGroup
}

func BuildServices(services []ServiceSchema) []ServiceData {
	out := []ServiceData{}
	var serviceErrors error

	for _, s := range services {
		serviceData, err := NewProtoService(s)
		serviceErrors = errors.Join(serviceErrors, indentErrors(fmt.Sprintf("Errors for the service schema %q", s.Resource.Name), err))
		out = append(out, serviceData)
	}

	if serviceErrors != nil {
		fmt.Printf("The following errors occurred:\n\n")
		log.Fatal(serviceErrors)
	}

	return out
}

func NewProtoService(s ServiceSchema) (ServiceData, error) {
	imports := make(Set)
	processedMessages := make(Set)

	messages := make([]MessageSchema, len(s.Messages))
	copy(messages, s.Messages)

	out := &ServiceData{ResourceName: s.Resource.Name, FileOptions: s.FileOptions, ServiceOptions: s.ServiceOptions, Imports: imports, Extensions: s.OptionExtensions, Enums: s.Enums}

	var messageErrors error

	processMessage := func(m MessageSchema) {
		var errAgg error

		message, err := NewProtoMessage(m, imports)
		errAgg = errors.Join(errAgg, err)
		out.Messages = append(out.Messages, message)
		processedMessages[m.Name] = present

		if errAgg != nil {
			messageErrors = errors.Join(messageErrors, indentErrors(fmt.Sprintf("Errors for the %s message schema", m.Name), errAgg))
		}
	}

	processMessage(s.Resource)

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
			return sortString(a, b)
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
