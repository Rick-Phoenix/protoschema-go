package schemabuilder

type FieldData map[string]*ServiceData

var UserSchema = ProtoMessageSchema{
	Fields: ProtoFieldsMap{
		"name": ProtoString(1),
	},
}

type ServicesMap map[string]ProtoServiceSchema

type ServicesData map[string]ProtoService

// Make something that reflects the db field names and types and checks if the messages are correct
func BuildFinalServicesMap(m ServicesMap) ServicesData {
	out := make(ServicesData)

	for resource, serviceSchema := range m {
		out[resource] = NewProtoService(resource, serviceSchema)
	}

	return out
}

var TablesData = ServicesMap{
	"User": ProtoServiceSchema{
		Get: &ServiceData{
			Request:  ProtoMessageSchema{},
			Response: ProtoMessageSchema{},
		},
	},
}
