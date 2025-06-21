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
	database, err := sql.Open("sqlite", "file:///home/rick/go-first/db/database.sqlite3?_time_format=sqlite")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	queries := gofirst.New(database)
	ctx := context.Background()

	user, err := queries.GetUser(ctx, 1)
	posts, err := queries.GetPostsFromUserId(ctx, 1)

	userData := UserWithPosts{User: user, Posts: posts}

	fmt.Printf("%+v", userData)

}
