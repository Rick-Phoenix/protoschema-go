package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Rick-Phoenix/gofirst/db/sqlgen"
)

type Store struct {
	db      sqlgen.DBTX
	Queries *sqlgen.Queries
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: sqlgen.New(db),
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
	tx, err := s.db.(*sql.DB).BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	qtx := s.Queries.WithTx(tx)
	user, err := qtx.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	posts, err := qtx.GetPostsFromUserId(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit read transaction: %w", err)
	}

	return &UserWithPosts{
		User: user, Posts: ToPointer(posts),
	}, nil
}
