-- name: GetUserWithPostsFromView :one
SELECT * FROM user_with_posts WHERE id = ?;

-- name: GetUser :one
SELECT * FROM users WHERE id = ?;

-- name: GetPostsFromUserId :many
SELECT * FROM posts WHERE author_id = ?;

-- name: CreateUser :one
INSERT INTO users (name) VALUES (?) Returning * ;
