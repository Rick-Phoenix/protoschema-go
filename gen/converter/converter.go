package converter


import (
  "github.com/Rick-Phoenix/gofirst/db"
  "github.com/Rick-Phoenix/gofirst/db/sqlgen"
  "github.com/Rick-Phoenix/gofirst/gen/myappv1"
  "google.golang.org/protobuf/types/known/timestamppb"
  )




func UserToUserMsg(User *db.UserWithPosts) *myappv1.User {
	if User == nil {
		return nil
	}
  CreatedAt := timestamppb.New(User.CreatedAt)
  return &myappv1.User{
    Id: User.Id,
    Name: User.Name,
    CreatedAt: CreatedAt,
    Posts: PostsToPostsMsg(User.Posts),
    
	}
}

func PostToPostMsg(Post *sqlgen.Post) *myappv1.Post {
	if Post == nil {
		return nil
	}
  CreatedAt := timestamppb.New(Post.CreatedAt)
  return &myappv1.Post{
    Id: Post.Id,
    Title: Post.Title,
    CreatedAt: CreatedAt,
    AuthorId: Post.AuthorId,
    SubredditId: Post.SubredditId,
    
	}
}

func PostsToPostsMsg(Post []*sqlgen.Post) []*myappv1.Post {
	out := make([]*myappv1.Post, len(Post))

	for _, v := range Post {
		out = append(out, PostToPostMsg(v))
	}

	return out
}
