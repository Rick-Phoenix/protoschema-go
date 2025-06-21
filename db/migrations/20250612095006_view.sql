-- migrate:up
CREATE VIEW user_with_posts AS
SELECT
    u.id,
    u.name,
    u.created_at,
    COALESCE(
        JSONB_GROUP_ARRAY(
            JSONB_OBJECT(
                'id', p.id,
                'title', p.title,
                'content', p.content,
                'created_at', p.created_at,
                'author_id', p.author_id,
                'subreddit_id', p.subreddit_id
            )
        ) FILTER (WHERE p.id IS NOT NULL),
        '[]'
    ) AS posts
FROM
    users AS u
LEFT JOIN
    posts AS p
    ON u.id = p.author_id
GROUP BY
    u.id, u.name, u.created_at;

INSERT INTO users (name) VALUES ("gianfranco");
INSERT INTO subreddits (name, creator_id) VALUES ("r/cats", 1);
INSERT INTO posts (title, author_id, subreddit_id) VALUES (
    "cats are neat eh?", 1, 1
);

-- migrate:down
DROP VIEW user_with_posts;
DELETE FROM users WHERE name = "gianfranco";
DELETE FROM subreddits WHERE id = 1;
DELETE FROM posts WHERE id = 1;
