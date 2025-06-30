

package converter


import (
  "github.com/Rick-Phoenix/gofirst/db"
  "github.com/Rick-Phoenix/gofirst/db/sqlgen"
  "github.com/Rick-Phoenix/gofirst/gen/myappv1"
  "google.golang.org/protobuf/types/known/timestamppb"
  )




func PostToPostMsg(Post *sqlgen.Post) *myappv1.Post {
	if Post == nil {
		return nil
	}
  CreatedAt := timestamppb.New(Post.CreatedAt)
  return &myappv1.Post{
    Id: Post.Id,
    Title: Post.Title,
    Content: Post.Content,
    CreatedAt: CreatedAt,
    AuthorId: Post.AuthorId,
    SubredditId: Post.SubredditId,
    
	}
}

func PostsToPostsMsg(Post []*sqlgen.Post) []*myappv1.Post {
	out := make([]*myappv1.Post, len(Post))

	for i, v := range Post {
    out[i] = PostToPostMsg(v)
	}

	return out
}
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

