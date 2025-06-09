CREATE TABLE IF NOT EXISTS "schema_migrations" (version varchar(255) primary key);
CREATE TABLE users (
-- sqlite: integer primary key will auto-increment by default
    id integer primary key,
    -- add autoincrement if you need strictly monotonically increasing ids
    -- that are never reused, even after deletes.
    name text not null unique,
    -- common sqlite way for timestamps
    created_at text default current_timestamp
);
CREATE TABLE subreddits (
    id integer primary key,
    name text not null unique, -- varchar(100) becomes text affinity in sqlite
    description text,
    created_at text default current_timestamp,
    creator_id integer, -- nullable
    foreign key (creator_id) references users (id) on delete set null
);
CREATE TABLE posts (
    id integer primary key,
    title text not null, -- varchar(300) becomes text affinity
    content text,
    created_at text default current_timestamp,
    author_id integer not null,
    subreddit_id integer not null,
    foreign key (author_id) references users (id) on delete cascade,
    foreign key (subreddit_id) references subreddits (id) on delete cascade
);
CREATE TABLE comments (
    id integer primary key,
    text_content text not null,
    created_at text default current_timestamp,
    author_id integer not null,
    post_id integer not null,
    parent_comment_id integer, -- nullable for top-level comments
    foreign key (author_id) references users (id) on delete cascade,
    foreign key (post_id) references posts (id) on delete cascade,
    foreign key (parent_comment_id) references comments (id) on delete cascade
);
CREATE TABLE user_subscriptions (
    user_id integer not null,
    subreddit_id integer not null,
    created_at text default current_timestamp,
    primary key (user_id, subreddit_id),
    foreign key (user_id) references users (id) on delete cascade,
    foreign key (subreddit_id) references subreddits (id) on delete cascade
);
-- Dbmate schema migrations
INSERT INTO "schema_migrations" (version) VALUES
  ('20250609101228'),
  ('20250609104808'),
  ('20250609135615'),
  ('20250609140445'),
  ('20250609153103');
