CREATE TABLE IF NOT EXISTS "schema_migrations" (version varchar(255) primary key);
CREATE TABLE users (
id integer primary key,
name text not null unique,
created_at datetime not null default (strftime ('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE TABLE subreddits (
id integer primary key,
name text not null unique,
description text,
created_at datetime not null default (strftime ('%Y-%m-%dT%H:%M:%fZ', 'now')),
creator_id integer,
foreign key (creator_id) references users (id) on delete set null
);
CREATE TABLE posts (
id integer primary key,
title text not null,
content text,
created_at datetime not null default (strftime ('%Y-%m-%dT%H:%M:%fZ', 'now')),
author_id integer not null,
subreddit_id integer not null,
foreign key (author_id) references users (id) on delete cascade,
foreign key (subreddit_id) references subreddits (id) on delete cascade
);
CREATE TABLE comments (
id integer primary key,
text_content text not null,
created_at datetime not null default (strftime ('%Y-%m-%dT%H:%M:%fZ', 'now')),
author_id integer not null,
post_id integer not null,
parent_comment_id integer,
foreign key (author_id) references users (id) on delete cascade,
foreign key (post_id) references posts (id) on delete cascade,
foreign key (parent_comment_id) references comments (id) on delete cascade
);
CREATE TABLE user_subscriptions (
user_id integer not null,
subreddit_id integer not null,
created_at datetime not null default (strftime ('%Y-%m-%dT%H:%M:%fZ', 'now')),
primary key (user_id, subreddit_id),
foreign key (user_id) references users (id) on delete cascade,
foreign key (subreddit_id) references subreddits (id) on delete cascade
);
-- Dbmate schema migrations
INSERT INTO "schema_migrations" (version) VALUES
  ('20250609140445'),
  ('20250612095006');
