-- name: GetUserWithPostsFromView :one
SELECT * FROM user_with_posts;

-- name: GetUser :one
SELECT * FROM users WHERE id = ?;

-- name: GetPostsFromUserId :many
SELECT * FROM posts WHERE author_id = ?;
