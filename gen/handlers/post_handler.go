package handlers


import (
  "github.com/Rick-Phoenix/gofirst/gen/myappv1"
  )



type PostService struct {
	Store *db.Store
}

func NewPostService(s *db.Store) *PostService {
	return &PostService{Store: s}
}






func (s *PostService) UpdatePost(
	ctx context.Context,
  req *connect.Request[myappv1.UpdatePostRequest],
) (*connect.Response[myappv1.google.protobuf.Empty], error) {

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

  return connect.NewResponse(&myappv1.google.protobuf.Empty{

	}), nil
}



func (s *PostService) GetPost(
	ctx context.Context,
  req *connect.Request[myappv1.GetPostRequest],
) (*connect.Response[myappv1.GetPostResponse], error) {

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

  return connect.NewResponse(&myappv1.GetPostResponse{

	}), nil
}


