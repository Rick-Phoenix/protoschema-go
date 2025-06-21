package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
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

	userWithPosts, err := queries.GetUserWithPostsFromView(ctx, 1)

	var posts []gofirst.Post
	err = json.Unmarshal(userWithPosts.Posts, &posts)

	userData := UserWithPosts{User: gofirst.User{ID: userWithPosts.ID, Name: userWithPosts.Name, CreatedAt: userWithPosts.CreatedAt}, Posts: posts}

	fmt.Printf("%+v", userData)

}
