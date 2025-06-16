package schemabuilder

type FieldData map[string]*ServiceData

var UserSchema = ProtoMessageSchema{
	Fields: ProtoFieldsMap{
		"name": ProtoString(1),
	},
}

var OverrideSchema = ExtendProtoMessage(UserSchema, &ProtoMessageSchema{
	Fields: ProtoFieldsMap{
		"name": ProtoString(2),
	},
})

type ServicesMap map[string]ProtoServiceSchema

type ServicesData map[string]ProtoService

// Make something that reflects the db field names and types and checks if the messages are correct
func BuildFinalServicesMap(m ServicesMap) ServicesData {
	out := make(ServicesData)

	for resource, serviceSchema := range m {
		out[resource] = NewProtoService(resource, serviceSchema, "myapp/v1")
	}

	return out
}

// Generate generic message type
var TablesData = ServicesMap{
	"User": ProtoServiceSchema{
		Resource: UserSchema,
		Get: &ServiceData{
			Request: *OverrideSchema,
			Response: ProtoMessageSchema{
				Reserved: []int{100, 101, 102},
				Fields: ProtoFieldsMap{
					"user": ExternalType(1, "User"),
					"createdAt": ProtoTimestamp(2).Required().CelField(CelFieldOpts{
						Id:         "test",
						Message:    "this is a test",
						Expression: "this = test",
					}),
				},
			},
		},
	},
}

var UserService = NewProtoService("User", TablesData["User"], "myapp/v1")
