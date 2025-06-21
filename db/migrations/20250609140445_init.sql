-- migrate:up

create table users (
id integer primary key,
name text not null unique,
created_at datetime not null default (strftime ('%Y-%m-%dT%H:%M:%fZ', 'now'))
) ;

create table subreddits (
id integer primary key,
name text not null unique,
description text,
created_at datetime not null default (strftime ('%Y-%m-%dT%H:%M:%fZ', 'now')),
creator_id integer,
foreign key (creator_id) references users (id) on delete set null
) ;

create table posts (
id integer primary key,
title text not null,
content text,
created_at datetime not null default (strftime ('%Y-%m-%dT%H:%M:%fZ', 'now')),
author_id integer not null,
subreddit_id integer not null,
foreign key (author_id) references users (id) on delete cascade,
foreign key (subreddit_id) references subreddits (id) on delete cascade
) ;

create table comments (
id integer primary key,
text_content text not null,
created_at datetime not null default (strftime ('%Y-%m-%dT%H:%M:%fZ', 'now')),
author_id integer not null,
post_id integer not null,
parent_comment_id integer,
foreign key (author_id) references users (id) on delete cascade,
foreign key (post_id) references posts (id) on delete cascade,
foreign key (parent_comment_id) references comments (id) on delete cascade
) ;

create table user_subscriptions (
user_id integer not null,
subreddit_id integer not null,
created_at datetime not null default (strftime ('%Y-%m-%dT%H:%M:%fZ', 'now')),
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
