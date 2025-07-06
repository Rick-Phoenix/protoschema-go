
package handlers


import (
  "github.com/Rick-Phoenix/protoschema/_test/db"
  "github.com/Rick-Phoenix/protoschema/gen/myappv1"
  )



type PostService struct {
  Queries *db.Queries
}

func NewPostService(s *db.Store) *PostService {
	return &PostService{Store: s}
}





func (s *PostService) GetPost(
	ctx context.Context,
  req *connect.Request[myappv1.GetPostRequest],
) (*connect.Response[myappv1.GetPostResponse], error) {

  
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

  return connect.NewResponse(&myappv1.GetPostResponse{
    
    
    Post: converter.PosttoPostMsg(post),
    
  }), nil
}



func (s *PostService) UpdatePost(
	ctx context.Context,
  req *connect.Request[myappv1.UpdatePostRequest],
) (*connect.Response[emptypb.Empty], error) {

  
post, err := s.Queries.UpdatePost(ctx, req.Msg.Get())

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

  return connect.NewResponse(&emptypb.Empty{
    
  }), nil
}



