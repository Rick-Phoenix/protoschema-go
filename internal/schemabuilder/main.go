package schemabuilder

type FieldData map[string]*ServiceData

// Defining an external type (like User) + integrating its importt directly
// Using separate generic resource messages which then get implemented by GetResponse

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
				Fields: ProtoFieldsMap{
					"user": ExternalType(1, "User"),
				},
			},
		},
	},
}
