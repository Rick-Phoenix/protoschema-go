package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/Rick-Phoenix/protoschema/_test/db"
	"github.com/Rick-Phoenix/protoschema/gen/converter"
	"github.com/Rick-Phoenix/protoschema/gen/myappv1"
	"google.golang.org/protobuf/types/known/emptypb"
	"modernc.org/sqlite"
)

type UserService struct {
	Queries *db.Queries
}

func NewUserService(s *db.Store) *UserService {
	return &UserService{Store: s}
}

func (s *UserService) GetUser(
	ctx context.Context,
	req *connect.Request[myappv1.GetUserRequest],
) (*connect.Response[myappv1.GetUserResponse], error) {
	user, err := s.Queries.GetUser(ctx, req.Msg.Get())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		} else {
			var sqliteErr *sqlite.Error
			if errors.As(err, &sqliteErr) {
				fmt.Printf("Sqlite error: %s\n", sqlite.ErrorCodeString[sqliteErr.Code()])
			} else {
				fmt.Printf("Unknown error: %s\n", err.Error())
			}
			return nil, connect.NewError(connect.CodeUnknown, err)
		}
	}

	return connect.NewResponse(&myappv1.GetUserResponse{
		User: converter.UserToUserMsg(user),
	}), nil
}

func (s *UserService) UpdateUser(
	ctx context.Context,
	req *connect.Request[myappv1.UpdateUserRequest],
) (*connect.Response[emptypb.Empty], error) {
	err := s.Queries.UpdateUser(ctx, req.Msg.Get())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		} else {
			var sqliteErr *sqlite.Error
			if errors.As(err, &sqliteErr) {
				fmt.Printf("Sqlite error: %s\n", sqlite.ErrorCodeString[sqliteErr.Code()])
			} else {
				fmt.Printf("Unknown error: %s\n", err.Error())
			}
			return nil, connect.NewError(connect.CodeUnknown, err)
		}
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}
