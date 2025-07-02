package protoschema

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type connectHandler struct {
	ServiceData
	Imports Set
}

func (p *ProtoPackage) genConnectHandler(f FileData) error {
	tmpl := p.tmpl

	for _, s := range f.Services {
		var handlerBuffer bytes.Buffer
		handlerData := connectHandler{Imports: Set{p.GoPackagePath: present}, ServiceData: s}
		if err := tmpl.ExecuteTemplate(&handlerBuffer, "connectHandler", handlerData); err != nil {
			return fmt.Errorf("Failed to execute template: %w", err)
		}

		handlerOut := filepath.Join("gen/handlers", strings.ToLower(s.Resource)+"_handler.go")

		if err := os.MkdirAll(filepath.Dir(handlerOut), 0755); err != nil {
			return err
		}

		if err := os.WriteFile(handlerOut, handlerBuffer.Bytes(), 0644); err != nil {
			return err
		}

		fmt.Printf("✅ Generated handler in %s\n", handlerOut)

	}

	return nil
}
