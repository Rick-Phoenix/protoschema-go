-- migrate:up
CREATE VIEW user_with_posts AS
SELECT
    u.id,
    u.name,
    u.created_at,
    COALESCE(
        JSON_GROUP_ARRAY(
            JSON_OBJECT(
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

-- migrate:down
DROP VIEW user_with_posts;
