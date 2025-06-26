package schemabuilder

var ProtoTypeMap = map[string]string{
	"string":                    "string",
	"bytes":                     "[]byte",
	"bool":                      "bool",
	"float":                     "float32",
	"double":                    "float64",
	"int32":                     "int32",
	"int64":                     "int64",
	"uint32":                    "uint32",
	"uint64":                    "uint64",
	"sint32":                    "int32",
	"sint64":                    "int64",
	"fixed32":                   "uint32",
	"fixed64":                   "uint64",
	"sfixed32":                  "int32",
	"sfixed64":                  "int64",
	"google.protobuf.Timestamp": "timestamppb.Timestamp",
	"google.protobuf.Duration":  "durationpb.Duration",
	"google.protobuf.FieldMask": "field_mask",
	"google.protobuf.Any":       "any",
}
