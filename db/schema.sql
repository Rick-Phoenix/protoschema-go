CREATE TABLE IF NOT EXISTS "schema_migrations" (
    version varchar(255) PRIMARY KEY
);
CREATE TABLE users (
    id integer PRIMARY KEY,
    name text NOT NULL UNIQUE,
    created_at text DEFAULT current_timestamp
);
CREATE TABLE subreddits (
    id integer PRIMARY KEY,
    name text NOT NULL UNIQUE,
    description text,
    created_at text DEFAULT current_timestamp,
    creator_id integer,
    FOREIGN KEY (creator_id) REFERENCES users (id) ON DELETE SET NULL
);
CREATE TABLE posts (
    id integer PRIMARY KEY,
    title text NOT NULL,
    content text,
    created_at text DEFAULT current_timestamp,
    author_id integer NOT NULL,
    subreddit_id integer NOT NULL,
    FOREIGN KEY (author_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (subreddit_id) REFERENCES subreddits (id) ON DELETE CASCADE
);
CREATE TABLE comments (
    id integer PRIMARY KEY,
    text_content text NOT NULL,
    created_at text DEFAULT current_timestamp,
    author_id integer NOT NULL,
    post_id integer NOT NULL,
    parent_comment_id integer,
    FOREIGN KEY (author_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE,
    FOREIGN KEY (parent_comment_id) REFERENCES comments (id) ON DELETE CASCADE
);
CREATE TABLE user_subscriptions (
    user_id integer NOT NULL,
    subreddit_id integer NOT NULL,
    created_at text DEFAULT current_timestamp,
    PRIMARY KEY (user_id, subreddit_id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (subreddit_id) REFERENCES subreddits (id) ON DELETE CASCADE
);
-- Dbmate schema migrations
INSERT INTO "schema_migrations" (version) VALUES
('20250609101228'),
('20250609104808'),
('20250609135615'),
('20250609140445'),
('20250609153103');
