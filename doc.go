// This package does the following:
// 1. It allows the definition of the contents of protobuf files in a declarative manner, as go structs.
// 2. It allows the user to define a specific model (usually a struct that represents a database item) and bind it to a protobuf message schema, so that if one of the two changes, this package will catch it and report the errors.
// 3. It automatically generates functions that convert structs of a specific type to their respective protobuf message types.
// 4. It exposes several hooks that allow the user to customize the logic for all of the above.
// 5. It provides a way to easily define protovalidate rules for protobuf messages, oneofs and fields, in a way that is typesafe and attempts to catch most erroneous definitions.
package protoschema
