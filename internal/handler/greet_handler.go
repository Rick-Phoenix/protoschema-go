// internal/handler/greet_handler.go
package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"buf.build/go/protovalidate"

	"connectrpc.com/connect"
	"google.golang.org/genproto/googleapis/rpc/errdetails"

	greetv1 "github.com/Rick-Phoenix/gofirst/gen/greet/v1"
	greetv1connect "github.com/Rick-Phoenix/gofirst/gen/greet/v1/greetv1connect"
)

// GreetServer (definition remains the same)
type GreetServer struct {
	validator protovalidate.Validator
	greetv1connect.UnimplementedGreetServiceHandler
}

// NewGreetServer (definition remains the same)
func NewGreetServer() (*GreetServer, error) {
	v, err := protovalidate.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create protovalidate validator: %w", err)
	}
	return &GreetServer{
		validator: v,
	}, nil
}

func (s *GreetServer) Greet(
	ctx context.Context,
	req *connect.Request[greetv1.DataRequest],
) (*connect.Response[greetv1.DataResponse], error) {
	if err := s.validator.Validate(req.Msg); err != nil {
		var pvValidationError *protovalidate.ValidationError
		if errors.As(err, &pvValidationError) {
			connErr := connect.NewError(connect.CodeInvalidArgument, errors.New("Validation failed. See details."))
			br := &errdetails.BadRequest{}

			for _, violationWrapper := range pvValidationError.Violations {
				protoViolation := violationWrapper.Proto // This is *validate.Violation
				if protoViolation != nil {
					var fieldPathStr string
					fieldPathMessage := protoViolation.GetField() // This is *validate.FieldPath

					if fieldPathMessage != nil && len(fieldPathMessage.GetElements()) > 0 {
						var segments []string
						for _, element := range fieldPathMessage.GetElements() {
							if element.HasFieldName() { // Check if 'field_name' is set
								segments = append(segments, element.GetFieldName())
							} else if element.HasIndex() { // Check if 'index' is set
								segments = append(segments, fmt.Sprintf("[%d]", element.GetIndex()))
							} else if element.HasStringKey() { // Check if 'string_key' is set
								segments = append(segments, element.GetStringKey())
							} else if element.HasIntKey() { // Check if 'int_key' is set
								segments = append(segments, fmt.Sprintf("[%d]", element.GetIntKey()))
							} else if element.HasBoolKey() { // Check if 'bool_key' is set
								segments = append(segments, fmt.Sprintf("%t", element.GetBoolKey()))
								// Add other key types if necessary (e.g., GetUintKey)
								// } else if element.HasUintKey() {
								// 	segments = append(segments, fmt.Sprintf("[%d]", element.GetUintKey()))
							} else {
								// Fallback if it's an unknown or unhandled element type
								// You could use element.String() for a verbose representation if needed
								segments = append(segments, "<unknown_element_kind>")
							}
						}
						fieldPathStr = strings.Join(segments, ".")
					}

					// Fallback if the path string is still empty
					if fieldPathStr == "" && fieldPathMessage != nil {
						fieldPathStr = fieldPathMessage.String() // Original verbose string from FieldPath
					} else if fieldPathStr == "" {
						fieldPathStr = "<unknown_field>"
					}

					br.FieldViolations = append(br.FieldViolations, &errdetails.BadRequest_FieldViolation{
						Field:       fieldPathStr,
						Description: protoViolation.GetMessage(),
					})
				}
			}

			detail, detailErr := connect.NewErrorDetail(br)
			if detailErr == nil {
				connErr.AddDetail(detail)
			} else {
				fmt.Printf("Error creating error detail: %v\n", detailErr)
			}
			return nil, connErr
		}
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("request processing error: %w", err))
	}

	name := req.Msg.GetName()
	age := req.Msg.GetAge()
	greeting := fmt.Sprintf("Hello, %s! You are %d years old, and your request is valid!", name, age)
	res := connect.NewResponse(&greetv1.DataResponse{
		Message: greeting,
	})
	return res, nil
}

var _ greetv1connect.GreetServiceHandler = (*GreetServer)(nil)
