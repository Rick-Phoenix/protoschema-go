package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Rick-Phoenix/gofirst/db/sqlgen"
)

type Store struct {
	db      sqlgen.DBTX
	queries *sqlgen.Queries
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		queries: sqlgen.New(db),
	}
}

type UserWithPosts struct {
	sqlgen.User
	Posts []*sqlgen.Post
}

func ToPointer[T any](s []T) []*T {
	out := make([]*T, len(s))
	for _, v := range s {
		out = append(out, &v)
	}

	return out
}

func (s *Store) GetUserWithPosts(ctx context.Context, userID int64) (*UserWithPosts, error) {
	user, err := s.queries.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	posts, err := s.queries.GetPostsFromUserId(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	return &UserWithPosts{
		User: user, Posts: ToPointer(posts),
	}, nil
}
