package handlers


import (
  "github.com/Rick-Phoenix/gofirst/gen/myappv1"
  )



type UserService struct {
	Store *db.Store
}

func NewUserService(s *db.Store) *UserService {
	return &UserService{Store: s}
}





func (s *UserService) GetUser(
	ctx context.Context,
  req *connect.Request[myappv1.GetUserRequest],
) (*connect.Response[myappv1.GetUserResponse], error) {

  resource, err := s.Store.method(ctx, params)
  if errors.Is(err, sql.ErrNoRows) {
    return nil, connect.NewError(connect.CodeNotFound, err)
  } else {
    var sqliteErr *sqlite.Error
    if errors.As(err, &sqliteErr) {
      switch sqliteErr.Code() {
      case sqlite3.SQLITE_CONSTRAINT:
        //
      }
    }
  }

  return connect.NewResponse(&myappv1.GetUserResponse{

	}), nil
}



func (s *UserService) UpdateUser(
	ctx context.Context,
  req *connect.Request[myappv1.UpdateUserRequest],
) (*connect.Response[emptypb.Empty], error) {

  resource, err := s.Store.method(ctx, params)
  if errors.Is(err, sql.ErrNoRows) {
    return nil, connect.NewError(connect.CodeNotFound, err)
  } else {
    var sqliteErr *sqlite.Error
    if errors.As(err, &sqliteErr) {
      switch sqliteErr.Code() {
      case sqlite3.SQLITE_CONSTRAINT:
        //
      }
    }
  }

  return connect.NewResponse(&emptypb.Empty{

	}), nil
}


