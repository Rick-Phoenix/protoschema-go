package protoschema

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/*
var templateFS embed.FS

// The function that processes the file schemas (and all the schemas inside them) and generates the proto files, while also calling the various hooks and the converter function.
// This should be called after all the elements of the proto package have been added with the various constructors.
func (p *ProtoPackage) Generate() error {
	filesData := p.BuildFiles()

	tmpl := p.tmpl

	for _, fileData := range filesData {

		outputFile := strings.ToLower(fileData.Name)
		outputPath := filepath.Join(p.protoOutputDir, outputFile)
		delete(fileData.Imports, filepath.Join(p.GetBasePath(), outputFile))

		var outputBuffer bytes.Buffer
		if err := tmpl.ExecuteTemplate(&outputBuffer, "protoFile", fileData); err != nil {
			return fmt.Errorf("Failed to execute template: %w", err)
		}

		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(outputPath, outputBuffer.Bytes(), 0644); err != nil {
			return err
		}

		fmt.Printf("✅ Successfully generated proto file at: %s\n", outputPath)

		_, err := exec.LookPath("buf")
		if err != nil {
			fmt.Println("Could not format the generated proto file. Is the buf cli in PATH?")
		} else {
			cmd := exec.Command("buf", "format", "-w", outputPath)

			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				fmt.Printf("Error while attempting to format the file %q: %s\n", outputPath, err.Error())
			}
		}

	}

	if p.converterFunc == nil {
		var outputBuffer bytes.Buffer
		if err := tmpl.ExecuteTemplate(&outputBuffer, "converter", p.converter); err != nil {
			fmt.Printf("Failed to execute template: %s", err.Error())
		}

		outputPath := filepath.Join(p.converterOutputDir, p.converterPackage+".go")

		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return err
		}

		if err := os.WriteFile(outputPath, outputBuffer.Bytes(), 0644); err != nil {
			fmt.Print(err)
		}

		cmd := exec.Command("gofmt", "-w", outputPath)
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			fmt.Printf("An error occurred while trying to format the converter file at %q:\n%s\n", outputPath, err.Error())
		}

		importCmd := exec.Command("goimports", "-w", outputPath)
		importCmd.Stderr = os.Stderr
		importErr := importCmd.Run()
		if importErr != nil {
			fmt.Printf("An error occurred while trying to call goimports for the file %q:\n%s\n", outputPath, importErr.Error())
		}

		fmt.Printf("✅ Successfully generated converter at: %s\n", outputPath)

	}

	return nil
}

var funcMap = template.FuncMap{
	"fmtOpt": func(o ProtoOption) string {
		opt, err := getProtoOption(o.Name, o.Value)
		if err != nil {
			fmt.Println(err.Error())
			return "error"
		}

		return "option " + opt
	},
	"join": func(e []string, sep string) string {
		str := ""

		for i, s := range e {
			str += fmt.Sprintf("%q", s)

			if i != len(e)-1 {
				str += ", "
			}
		}

		return str
	},
	"joinInt":   joinIntSlice,
	"joinInt32": joinInt32Slice,
	"joinUint":  joinUintSlice,
	"joinRange": func(r []Range) string {
		str := ""

		for i, v := range r {
			str += fmt.Sprintf("%d to %d", v[0], v[1])
			if i != len(r)-1 {
				str += ", "
			}
		}

		return str
	},
	"dec": func(i int) int {
		return i - 1
	},
	"ternary": func(condition bool, trueVal any, falseVal any) any {
		if condition {
			return trueVal
		} else {
			return falseVal
		}
	},
	"keyword": func(isOptional bool, isRepeated bool) string {
		if isOptional {
			return "optional "
		} else if isRepeated {
			return "repeated "
		}

		return ""
	},
	"serviceSuffix": addServiceSuffix,
	"setContains": func(set Set, key string) bool {
		_, present := set[key]
		return present
	},
	"getProtoType": func(f FieldData, protoPackage *ProtoPackage) string {
		if f.MessageRef != nil {
			return f.MessageRef.GetFullName(protoPackage)
		}

		if f.EnumRef != nil {
			return f.EnumRef.GetFullName(protoPackage)
		}

		return f.ProtoType
	},
}
