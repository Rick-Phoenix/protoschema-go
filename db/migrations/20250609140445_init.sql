-- migrate:up

-- users table
create table users (
-- sqlite: integer primary key will auto-increment by default
id integer primary key,
-- add autoincrement if you need strictly monotonically increasing ids
-- that are never reused, even after deletes.
name text not null unique,
-- common sqlite way for timestamps
created_at datetime not null default (strftime ('%Y-%m-%dT%H:%M:%fZ',
'now'))
) ;

-- subreddits table
create table subreddits (
id integer primary key,
name text not null unique, -- varchar(100) becomes text affinity in sqlite
description text,
created_at text not null default current_timestamp,
creator_id integer, -- nullable
foreign key (creator_id) references users (id) on delete set null
) ;

-- posts table
create table posts (
id integer primary key,
title text not null, -- varchar(300) becomes text affinity
content text,
created_at text not null default current_timestamp,
author_id integer not null,
subreddit_id integer not null,
foreign key (author_id) references users (id) on delete cascade,
foreign key (subreddit_id) references subreddits (id) on delete cascade
) ;

-- comments table
create table comments (
id integer primary key,
text_content text not null,
created_at text default current_timestamp,
author_id integer not null,
post_id integer not null,
parent_comment_id integer, -- nullable for top-level comments
foreign key (author_id) references users (id) on delete cascade,
foreign key (post_id) references posts (id) on delete cascade,
foreign key (parent_comment_id) references comments (id) on delete cascade
) ;

-- user subscriptions table (many-to-many between users and subreddits)
create table user_subscriptions (
user_id integer not null,
subreddit_id integer not null,
created_at text not null default current_timestamp,
primary key (user_id, subreddit_id),
foreign key (user_id) references users (id) on delete cascade,
foreign key (subreddit_id) references subreddits (id) on delete cascade
) ;


-- migrate:down
drop table if exists user_subscriptions ;
drop table if exists comments ;
drop table if exists posts ;
drop table if exists subreddits ;
drop table if exists users ;
