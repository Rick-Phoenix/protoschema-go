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

func main() {
	// 1. Open the database
	database, err := sql.Open("sqlite", "db/database.sqlite3")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	queries := gofirst.New(database)
	ctx := context.Background()

	// 4. Fetch the user with all their posts using the view
	userWithPosts, err := queries.GetUserWithPostsFromView(ctx)
	if err != nil {
		log.Fatalf("Failed to get user with posts from view: %v", err)
	}

	var data []gofirst.Post
	err = json.Unmarshal([]byte(userWithPosts.Posts), &data)

	fmt.Printf("%+v", data)
}
