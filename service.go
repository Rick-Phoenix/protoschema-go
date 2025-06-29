package schemabuilder

import (
	"maps"
	"slices"
	"strings"
)

type HandlerData struct {
	Name     string
	Request  MessageSchema
	Response MessageSchema
}

type HandlersMap map[string]Handler

type Handler struct {
	Request  MessageSchema
	Response MessageSchema
}

type ServiceData struct {
	Resource string
	Options  []ProtoOption
	Handlers []HandlerData
}

type ServiceSchema struct {
	Resource string
	Handlers HandlersMap
	Options  []ProtoOption
}

func (s ServiceSchema) Build() ServiceData {
	out := ServiceData{
		Resource: s.Resource, Options: s.Options,
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

		out.Handlers = append(out.Handlers, HandlerData{Name: name, Request: h.Request, Response: h.Response})
	}

	return out
}
