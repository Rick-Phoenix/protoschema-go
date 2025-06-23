-- name: GetUser :one
SELECT
    *
FROM
    users
WHERE
    id = ?;

-- name: GetPostsFromUserId :many
SELECT
    *
FROM
    posts
WHERE
    author_id = ?;

-- name: CreateUser :one
INSERT INTO
    users (name)
VALUES
    (?)
RETURNING
    *;

-- name: PostWithUser :one
SELECT
    posts.*,
    sqlc.embed(u) AS author
FROM
    posts
    JOIN users u ON posts.author_id = users.id
WHERE
    users.id = ?;
