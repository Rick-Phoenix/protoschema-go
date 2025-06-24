package schemabuilder_test

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/Rick-Phoenix/gofirst/internal/schemabuilder"
	"github.com/bufbuild/protocompile/parser"
	"github.com/bufbuild/protocompile/reporter"
	"google.golang.org/protobuf/types/descriptorpb"
)

type FieldData struct {
	Name            string
	Number          int32
	Repeated        bool
	Optional        bool
	TypeName        string
	Options         map[string]ProtoOption
	RepeatedOptions []ProtoOption
}

type EnumMember struct {
	Name   string
	Number int32
}

type EnumData struct {
	Name            string
	Members         []EnumMember
	Options         map[string]ProtoOption
	RepeatedOptions []ProtoOption
	ReservedNames   []string
	ReservedNumbers []int32
	ReservedRanges  []schemabuilder.Range
}

type ProtoOption struct {
	Name  string
	Value any
}

type OneofData struct {
	Name            string
	Options         map[string]ProtoOption
	RepeatedOptions []ProtoOption
}

type MessageData struct {
	Name            string
	Fields          map[string]FieldData
	Enums           map[string]EnumData
	Oneofs          map[string]OneofData
	ReservedNames   []string
	ReservedNumbers []int32
	ReservedRanges  []schemabuilder.Range
	Options         map[string]ProtoOption
	RepeatedOptions []ProtoOption
}

type ServiceData struct {
	Name            string
	Methods         []MethodData
	Options         map[string]ProtoOption
	RepeatedOptions []ProtoOption
}

type MethodData struct {
	Name       string
	InputType  string
	OutputType string
}

type FileData struct {
	Messages   map[string]MessageData
	Enums      map[string]EnumData
	Services   map[string]ServiceData
	Extensions map[string]ExtensionData
}

type ExtensionData struct {
	Extendee string
	Fields   map[string]FieldData
}

func TestFirst(t *testing.T) {
	service, err := schemabuilder.NewProtoService("User", schemabuilder.UserService, "myapp/v1")
	if err != nil {
		log.Fatal(err)
	}

	tmpDir := t.TempDir()
	err = schemabuilder.GenerateProtoFile(service, schemabuilder.Options{ProtoRoot: tmpDir})
	if err != nil {
		log.Fatal(err)
	}

	filePath := path.Join(tmpDir, "myapp/v1", "user.proto")

	data := ParseProtoFile(filePath)

	fmt.Printf("DEBUG: %+v\n", data.Extensions)
}

func ParseProtoFile(filePath string) FileData {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open generated proto file: %v", err)
	}
	defer file.Close()
	errRep := reporter.NewReporter(func(err reporter.ErrorWithPos) error {
		return err
	}, func(reporter.ErrorWithPos) {})

	handler := reporter.NewHandler(errRep)
	ast, err := parser.Parse(filePath, file, handler)
	if err != nil {
		fmt.Print(err.Error())
	}

	data, err := parser.ResultFromAST(ast, false, handler)
	if err != nil {
		fmt.Print(err.Error())
	}

	desc := data.FileDescriptorProto()

	msgsMap := ExtractMessages(desc.GetMessageType())
	enumsMap := ExtractEnums(desc.GetEnumType())
	servicesMap := ExtractServices(desc.GetService())
	extensions := ExtractExtensions(desc.GetExtension())

	fileData := FileData{Messages: msgsMap, Enums: enumsMap, Services: servicesMap, Extensions: extensions}

	return fileData
}

func ExtractOpts(opts []*descriptorpb.UninterpretedOption) (map[string]ProtoOption, []ProtoOption) {
	optionsMap := make(map[string]ProtoOption)
	repeatedOptsMap := make(map[string]struct{})
	repeatedOpts := []ProtoOption{}
	for _, o := range opts {
		name := ""
		var value any

		for i, part := range o.GetName() {
			if i != 0 {
				name += "."
			}
			name += part.GetNamePart()
		}

		if o.StringValue != nil {
			value = o.String()
		} else if o.PositiveIntValue != nil {
			value = o.GetPositiveIntValue()
		} else if o.NegativeIntValue != nil {
			value = o.GetNegativeIntValue()
		} else if o.DoubleValue != nil {
			value = o.GetDoubleValue()
		} else if o.IdentifierValue != nil {
			strVal := o.GetIdentifierValue()

			strBool, err := strconv.ParseBool(strVal)
			if err != nil {
				value = o.String()
			} else {
				value = strBool
			}

		} else if o.AggregateValue != nil {
			value = o.String()
		}

		optsData := ProtoOption{
			Name: name, Value: value,
		}

		if repeatedOption, exists := optionsMap[name]; exists {
			if _, exists := repeatedOptsMap[name]; !exists {
				repeatedOptsMap[name] = struct{}{}
				repeatedOpts = append(repeatedOpts, repeatedOption)
			}
			repeatedOpts = append(repeatedOpts, optsData)
		} else {
			optionsMap[name] = optsData
		}

	}

	return optionsMap, repeatedOpts
}

func ExtractReservedNrs(ranges []*descriptorpb.DescriptorProto_ReservedRange) ([]int32, []schemabuilder.Range) {
	var reservedNumbers []int32
	var reservedRanges []schemabuilder.Range
	for _, r := range ranges {
		if r.Start != nil && r.End != nil {
			start := *r.Start
			end := *r.End

			if end-start == 1 {
				reservedNumbers = append(reservedNumbers, start)
			} else {
				reservedRanges = append(reservedRanges, schemabuilder.Range{start, end})
			}
		}
	}

	return reservedNumbers, reservedRanges
}

func ExtractEnumReservedNrs(ranges []*descriptorpb.EnumDescriptorProto_EnumReservedRange) ([]int32, []schemabuilder.Range) {
	var reservedNumbers []int32
	var reservedRanges []schemabuilder.Range
	for _, r := range ranges {
		if r.Start != nil && r.End != nil {
			start := *r.Start
			end := *r.End

			if end-start == 1 {
				reservedNumbers = append(reservedNumbers, start)
			} else {
				reservedRanges = append(reservedRanges, schemabuilder.Range{start, end})
			}
		}
	}

	return reservedNumbers, reservedRanges
}

func ExtractEnums(enums []*descriptorpb.EnumDescriptorProto) map[string]EnumData {
	data := make(map[string]EnumData)

	for _, e := range enums {
		var eData EnumData
		eData.Name = e.GetName()

		opts, repOpts := ExtractOpts(e.GetOptions().GetUninterpretedOption())

		eData.Options = opts
		eData.RepeatedOptions = repOpts

		for _, member := range e.GetValue() {
			var memData EnumMember
			memData.Name = member.GetName()
			memData.Number = member.GetNumber()
			eData.Members = append(eData.Members, memData)
		}

		eData.ReservedNames = e.GetReservedName()

		resNrs, resRanges := ExtractEnumReservedNrs(e.GetReservedRange())
		eData.ReservedNumbers = resNrs
		eData.ReservedRanges = resRanges
		data[eData.Name] = eData
	}

	return data
}

func ExtractServices(services []*descriptorpb.ServiceDescriptorProto) map[string]ServiceData {
	data := make(map[string]ServiceData)

	for _, serv := range services {
		var service ServiceData

		service.Name = serv.GetName()
		opts, repOpts := ExtractOpts(serv.GetOptions().GetUninterpretedOption())

		service.Options = opts
		service.RepeatedOptions = repOpts
		for _, m := range serv.GetMethod() {
			var method MethodData

			method.Name = m.GetName()
			method.InputType = m.GetInputType()
			method.OutputType = m.GetOutputType()

			service.Methods = append(service.Methods, method)
		}

		data[service.Name] = service
	}

	return data
}

func ExtractFields(fields []*descriptorpb.FieldDescriptorProto) map[string]FieldData {
	fieldsMap := make(map[string]FieldData)
	for _, f := range fields {
		rawOpts := f.GetOptions().GetUninterpretedOption()

		opts, repeatedOpts := ExtractOpts(rawOpts)

		fieldData := FieldData{
			Number: f.GetNumber(), Optional: f.GetProto3Optional(), Repeated: f.GetLabel().String() == "LABEL_REPEATED",
			TypeName: f.GetTypeName(), Name: f.GetName(), Options: opts, RepeatedOptions: repeatedOpts,
		}

		fieldsMap[f.GetName()] = fieldData

	}

	return fieldsMap
}

func ExtractExtensions(exts []*descriptorpb.FieldDescriptorProto) map[string]ExtensionData {
	data := make(map[string]ExtensionData)

	for _, f := range exts {
		var fieldData FieldData

		extendee := f.GetExtendee()

		if _, exists := data[extendee]; !exists {
			data[extendee] = ExtensionData{Extendee: extendee, Fields: make(map[string]FieldData)}
		}

		opts, repOpts := ExtractOpts(f.GetOptions().GetUninterpretedOption())
		fieldData.Options = opts
		fieldData.RepeatedOptions = repOpts
		fieldData.TypeName = strings.ToLower(strings.TrimPrefix(f.GetType().String(), "TYPE_"))
		fieldData.Name = f.GetName()

		data[extendee].Fields[fieldData.Name] = fieldData
	}

	return data
}

func ExtractMessages(messages []*descriptorpb.DescriptorProto) map[string]MessageData {
	msgsMap := make(map[string]MessageData)
	for _, m := range messages {
		fieldsMap := ExtractFields(m.GetField())

		oneofs := ExtractOneofs(m.GetOneofDecl())
		enumData := ExtractEnums(m.GetEnumType())

		opts, repOpts := ExtractOpts(m.GetOptions().GetUninterpretedOption())

		resNrs, resRanges := ExtractReservedNrs(m.GetReservedRange())

		msgData := MessageData{
			Name: m.GetName(), Fields: fieldsMap, ReservedNames: m.GetReservedName(), Options: opts, RepeatedOptions: repOpts, ReservedNumbers: resNrs, ReservedRanges: resRanges, Enums: enumData, Oneofs: oneofs,
		}

		msgsMap[m.GetName()] = msgData
	}

	return msgsMap
}

func ExtractOneofs(oneofs []*descriptorpb.OneofDescriptorProto) map[string]OneofData {
	data := make(map[string]OneofData)

	for _, o := range oneofs {
		opts, repOpts := ExtractOpts(o.GetOptions().GetUninterpretedOption())
		data[o.GetName()] = OneofData{Name: o.GetName(), Options: opts, RepeatedOptions: repOpts}
	}

	return data
}
