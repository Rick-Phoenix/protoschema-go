package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log"

	gofirst "github.com/Rick-Phoenix/gofirst/db/queries/gen"
	_ "modernc.org/sqlite"
)

type UserWithPosts struct {
	gofirst.User
	Posts []gofirst.Post
}

func main() {
	// 1. Open the database
	database, err := sql.Open("sqlite", "db/database.sqlite3")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	queries := gofirst.New(database)
	ctx := context.Background()

	user, err := queries.GetUsers(ctx, 1)

	posts, err := queries.GetPostsFromUserId(ctx, 1)

	userWithPosts := UserWithPosts{User: user, Posts: posts}

	fmt.Printf("%+v", userWithPosts)

}
