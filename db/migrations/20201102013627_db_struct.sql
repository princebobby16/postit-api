
-- +goose Up
CREATE SCHEMA postit;

CREATE TABLE IF NOT EXISTS postit.post
(
    post_id uuid UNIQUE NOT NULL,
    facebook_post_id character varying(200),
    post_message text NOT NULL,
    post_image bytea,
    image_extension character varying(200)[],
    hash_tags text[],
    post_status boolean NOT NULL,
    post_priority boolean NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (post_id)
);

CREATE TABLE IF NOT EXISTS postit.schedule
(
    schedule_id uuid UNIQUE NOT NULL,
    schedule_title character varying(200),
    post_to_feed boolean NOT NULL,
    schedule_from timestamp with time zone NOT NULL,
    schedule_to timestamp with time zone NOT NULL,
    post_ids character varying(200)[] NOT NULL,
    duration_per_post float NOT NULL,
    is_due boolean,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (schedule_id)
);

CREATE TABLE IF NOT EXISTS postit.scheduled_post
(
    scheduled_post_id uuid NOT NULL,
    post_id uuid NOT NULL,
    facebook_post_id character varying(200),
    post_message text NOT NULL,
    post_image bytea,
    image_extension character varying(200)[],
    hash_tags text[],
    post_status boolean NOT NULL,
    post_priority boolean NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS postit.application_info
(
    application_uuid uuid UNIQUE NOT NULL,
    application_name character varying (200) NOT NULL,
    application_id character varying(200) UNIQUE NOT NULL,
    application_secret character varying (200) NOT NULL,
    application_url character varying (200) NOT NULL,
    user_access_token text NOT NULL,
    expires_in integer NOT NULL,
    user_name character varying (200) UNIQUE NOT NULL,
    user_id character varying (200) UNIQUE NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(application_uuid)
);

-- +goose StatementBegin
create or replace function postit.change_post_status_in_post_table_aut() returns trigger as $$
DECLARE
    postId uuid = OLD.post_id;
    facebookId character varying(100) = OLD.facebook_post_id;
BEGIN
    UPDATE postit.post SET post_status = true AND facebook_post_id = facebookId WHERE post_id = postId;
    return old;
END;
$$ language plpgsql;

DROP TRIGGER IF EXISTS post_aut ON postit.scheduled_post;
CREATE TRIGGER post_aut AFTER INSERT ON postit.scheduled_post FOR EACH ROW EXECUTE PROCEDURE postit.change_post_status_in_post_table_aut();
-- +goose StatementEnd

-- +goose StatementBegin
create or replace function postit.store_schedule_data_in_scheduled_post_table_ait() returns trigger as $$

DECLARE
--     schedule vars
    iD           uuid := NEW.schedule_id;
    scheduleId   uuid;
    scheduleFrom timestamp;
    scheduleTo   timestamp;
    postList     varchar(200)[];

--     Post vars
    postId       uuid;
    postMessage  text;
    postImage    bytea;
    imageExtension character varying (200)[];
    hashTags     text[];
    postStatus boolean;
    postPriority boolean;
    facebookPostId character varying (200);

BEGIN

--     Get the schedule data
    SELECT schedule_id, schedule_from, schedule_to, post_ids INTO scheduleId, scheduleFrom, scheduleTo, postList FROM postit.schedule WHERE schedule_id = iD;
--     Loop through the post array retrieved from the schedule table to get the post ids
    FOREACH postId IN ARRAY postList
    LOOP
--      Use the post ids to retrieve the post info from the post table
        SELECT facebook_post_id, post_message, post_image, image_extension, hash_tags, post_priority, post_status INTO facebookPostId, postMessage, postImage, imageExtension, hashTags, postPriority, postStatus FROM postit.post WHERE post_id = postId;

--      Store it in the scheduled data table
        INSERT INTO postit.scheduled_post(scheduled_post_id, post_id, facebook_post_id, post_message, post_image, image_extension, hash_tags, post_status, post_priority) VALUES (scheduleId, postId, facebookPostId, postMessage, postImage, imageExtension, hashTags, postStatus, postPriority);

    END LOOP;

    RETURN NEW;

END;
$$ language plpgsql;

DROP TRIGGER IF EXISTS schedule_ait ON postit.schedule;
CREATE TRIGGER schedule_ait AFTER INSERT ON postit.schedule FOR EACH ROW EXECUTE PROCEDURE postit.store_schedule_data_in_scheduled_post_table_ait();
-- +goose StatementEnd

-- +goose StatementBegin
create or replace function postit.delete_posts_from_scheduled_post_table_bdt() returns trigger as $$

DECLARE

    schedule_id uuid = OLD.schedule_id;

BEGIN

    DELETE FROM postit.scheduled_post WHERE scheduled_post_id = schedule_id;

    RETURN OLD;

END;
$$ language plpgsql;

DROP TRIGGER IF EXISTS schedule_bdt ON postit.schedule;
CREATE TRIGGER schedule_bdt BEFORE DELETE ON postit.schedule FOR EACH ROW EXECUTE PROCEDURE postit.delete_posts_from_scheduled_post_table_bdt();
-- +goose StatementEnd

-- SQL in section 'Up' is executed when this migration is applied

-- +goose Down
DROP TABLE IF EXISTS postit.post;
DROP TABLE IF EXISTS postit.current_schedule;
DROP TABLE IF EXISTS postit.scheduled_post;
DROP TABLE IF EXISTS postit.application_info;
DROP TABLE IF EXISTS postit.schedule;
DROP SCHEMA IF EXISTS postit CASCADE;
-- SQL section 'Down' is executed when this migration is rolled back