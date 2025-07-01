package protoschema

import (
	"maps"
	"slices"
	"strings"
)

// Function that receives the service data after its schema has been processed.
type ServiceHook func(s ServiceData)

type HandlerData struct {
	Name     string
	Request  *MessageSchema
	Response *MessageSchema
}

// Maps handlers to their names.
type HandlersMap map[string]Handler

type Handler struct {
	Request  *MessageSchema
	Response *MessageSchema
}

// The output struct of the schema after it has been processed. Gets passed as an argument to the ServiceHook.
type ServiceData struct {
	Resource string
	File     *FileSchema
	Package  *ProtoPackage
	Options  []ProtoOption
	Handlers []HandlerData
	Metadata map[string]any
}

// The schema for a proto service. It should be created with the constructor from a FileSchema instance, as that populates the File and Package fields automatically.
type ServiceSchema struct {
	// The name of the service's resource. It will be joined with the "Service" suffix to create the service name. (i.e. "User" -> "UserService")
	Resource string
	// The FileSchema to which this service belongs. Set automatically when using the constructor.
	File *FileSchema
	// The ProtoPackage to which this service belongs. Set automatically when using the constructor.
	Package *ProtoPackage
	// A map of handlers. These correspond to the "rpc" directives in a proto file.
	Handlers HandlersMap
	Options  []ProtoOption
	// Schema-specific ServiceHook. If this is unset, and the service was created with the constructor, it defaults to the package-level ServiceHook. Otherwise, it overrides it.
	Hook ServiceHook
	// A map to store custom metadata to use in the hook. This gets passed directly to ServiceData instance.
	Metadata map[string]any
}

func (s *ServiceSchema) build(imports Set) ServiceData {
	out := ServiceData{
		Resource: s.Resource, Options: s.Options, Metadata: s.Metadata,
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

		handlerData := HandlerData{
			Name:     name,
			Request:  h.Request,
			Response: h.Response,
		}

		for _, v := range []*MessageSchema{h.Request, h.Response} {
			if v != nil && v.File != s.File && v.ImportPath != "" {
				imports[v.ImportPath] = present
			}
		}
		out.Handlers = append(out.Handlers, handlerData)
	}

	if s.Hook != nil {
		s.Hook(out)
	}

	return out
}
