-- +goose Up
CREATE SCHEMA postit;

CREATE TABLE IF NOT EXISTS postit.post
(
    post_id          uuid UNIQUE              NOT NULL,
    facebook_post_id character varying(200),
    facebook_user_id character varying(200),
    post_message     text                     NOT NULL,
    post_images      bytea[],
    image_paths      character varying(200)[],
    hash_tags        text[],
    post_fb_status   boolean,
    post_tw_status   boolean,
    post_li_status   boolean,
    scheduled        boolean,
    post_priority    boolean                  NOT NULL,
    created_at       timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (post_id)
);

CREATE TABLE IF NOT EXISTS postit.schedule
(
    schedule_id       uuid UNIQUE              NOT NULL,
    schedule_title    character varying(200),
    post_to_feed      boolean                  NOT NULL,
    schedule_from     timestamp with time zone NOT NULL,
    schedule_to       timestamp with time zone NOT NULL,
    post_ids          character varying(200)[] NOT NULL,
    duration_per_post float                    NOT NULL,
    facebook          character varying(200)[] NOT NULL,
    twitter           character varying(200)[] NOT NULL,
    linked_in         character varying(200)[] NOT NULL,
    is_due            boolean,
    created_at        timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (schedule_id)
);

CREATE TABLE IF NOT EXISTS postit.application_info
(
    application_uuid   uuid UNIQUE                   NOT NULL,
    application_name   character varying(200)        NOT NULL,
    application_id     character varying(200) UNIQUE NOT NULL,
    application_secret character varying(200)        NOT NULL,
    application_url    character varying(200)        NOT NULL,
    user_access_token  text                          NOT NULL,
    expires_in         integer                       NOT NULL,
    user_name          character varying(200) UNIQUE NOT NULL,
    user_id            character varying(200) UNIQUE NOT NULL,
    created_at         timestamp with time zone      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         timestamp with time zone      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (application_uuid)
);

CREATE TABLE IF NOT EXISTS general_schedule_table
(
    schedule_id       uuid UNIQUE              NOT NULL,
    posted_fb_ids     character varying(200)[],
    posted_tw_ids     character varying(200)[],
    posted_li_ids     character varying(200)[],
    tenant_namespace  character varying(200)   NOT NULL,
    duration_per_post float                    NOT NULL,
    created_at        timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (schedule_id)
);

-- SQL in section 'Up' is executed when this migration is applied

-- +goose Down
DROP TABLE IF EXISTS general_schedule_table;
DROP TABLE IF EXISTS postit.post;
DROP TABLE IF EXISTS postit.application_info;
DROP TABLE IF EXISTS postit.schedule;
DROP SCHEMA IF EXISTS postit CASCADE;
-- SQL section 'Down' is executed when this migration is rolled back