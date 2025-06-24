package schemabuilder_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
)

type FieldData struct {
	Name            string
	Number          int32
	Repeated        bool
	Optional        bool
	TypeName        string
	Deprecated      bool
	Options         map[string]FieldOption
	RepeatedOptions []FieldOption
}

type FieldOption struct {
	Name             string
	IdentifierValue  *string
	PositiveIntValue *uint64
	NegativeIntValue *int64
	DoubleValue      *float64
	StringValue      []byte
	AggregateValue   *string
}

type MessageData struct {
	Name   string
	Fields map[string]FieldData
}

func TestFirst(t *testing.T) {
	filePath := "/home/rick/go-first/gen/proto/myapp/v1/user.proto"
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open generated proto file: %v", err)
	}
	defer file.Close()
	errRep := reporter.NewReporter(func(err reporter.ErrorWithPos) error {
		return err
	}, func(reporter.ErrorWithPos) {})

	handler := reporter.NewHandler(errRep)
	ast, err := parser.Parse("/home/rick/go-first/gen/proto/myapp/v1/user.proto", file, handler)
	if err != nil {
		fmt.Print(err.Error())
	}

	data, err := parser.ResultFromAST(ast, false, handler)
	if err != nil {
		fmt.Print(err.Error())
	}

	desc := data.FileDescriptorProto()

	flatmsgs := desc.GetMessageType()

	msgsMap := make(map[string]MessageData)

	for _, m := range flatmsgs {
		fieldsMap := make(map[string]FieldData)
		msgData := MessageData{
			Name: m.GetName(), Fields: fieldsMap,
		}

		fmt.Printf("Data for message %s:\n", m.GetName())
		fields := m.GetField()

		for _, f := range fields {
			optionsMap := make(map[string]FieldOption)
			data := FieldData{
				Number: f.GetNumber(), Optional: f.GetProto3Optional(), Repeated: f.GetLabel().String() == "LABEL_REPEATED",
				TypeName: f.GetTypeName(), Name: f.GetName(), Options: optionsMap,
			}
			opts := f.GetOptions().GetUninterpretedOption()

			for _, o := range opts {
				name := ""

				for i, part := range o.GetName() {
					if i != 0 {
						name += "."
					}
					name += part.GetNamePart()
				}

				optsData := FieldOption{
					Name: name, IdentifierValue: o.IdentifierValue, StringValue: o.StringValue, AggregateValue: o.AggregateValue,
					PositiveIntValue: o.PositiveIntValue, NegativeIntValue: o.NegativeIntValue, DoubleValue: o.DoubleValue,
				}

				if _, exists := optionsMap[name]; exists {
					data.RepeatedOptions = append(data.RepeatedOptions, optsData)
				} else {
					data.Options[name] = optsData
				}
			}

			fieldsMap[f.GetName()] = data

		}

		msgsMap[m.GetName()] = msgData
	}
}
