BEGIN;

SET TIME ZONE 'Europe/London';

CREATE TABLE IF NOT EXISTS usr
(
    nickname    VARCHAR(128)        NOT NULL UNIQUE,
    fullname    VARCHAR(256)        NOT NULL,
    about       TEXT                DEFAULT NULL,
    email       VARCHAR(128)        UNIQUE,

    id          SERIAL              PRIMARY KEY
);

CREATE INDEX ON usr USING HASH ((LOWER(nickname)));

CREATE TABLE IF NOT EXISTS forum
(
    id          SERIAL              PRIMARY KEY,
    title       VARCHAR(1024)       NOT NULL,
    usr         VARCHAR(128)        NOT NULL REFERENCES usr (nickname) ON DELETE CASCADE,
    slug        VARCHAR(128)        NOT NULL UNIQUE,
    posts       BIGINT              DEFAULT 0,
    threads     INTEGER             DEFAULT 0,

    usr_id      INTEGER             NOT NULL REFERENCES usr (id) ON DELETE CASCADE
);

CREATE INDEX ON forum USING HASH ((LOWER(slug)));

CREATE TABLE IF NOT EXISTS thread
(
    id          SERIAL              PRIMARY KEY,
    title       VARCHAR(1024)       NOT NULL,
    author      VARCHAR(128)        REFERENCES usr (nickname) ON DELETE CASCADE,
    forum       VARCHAR(128)        REFERENCES forum (slug) ON DELETE CASCADE,
    message     TEXT                DEFAULT NULL,
    votes       INTEGER             DEFAULT 0,
    slug        VARCHAR(256)        UNIQUE,
    created     TIMESTAMP           DEFAULT current_timestamp,

    author_id   INTEGER             NOT NULL REFERENCES usr (id) ON DELETE CASCADE,
    forum_id    INTEGER             NOT NULL REFERENCES forum (id) ON DELETE CASCADE
);

CREATE INDEX ON thread USING HASH ((LOWER(slug)));
CREATE INDEX ON thread (created);

CREATE TABLE IF NOT EXISTS post
(
    id          BIGSERIAL           PRIMARY KEY,
    parent      BIGINT              DEFAULT 0,
    author      VARCHAR(128)        REFERENCES usr (nickname) ON DELETE CASCADE,
    message     TEXT                NOT NULL,
    isEdited    BOOLEAN             DEFAULT FALSE,
    forum       VARCHAR(128)        REFERENCES forum (slug) ON DELETE CASCADE,
    thread_id   INTEGER             REFERENCES thread (id) ON DELETE CASCADE,
    created     TIMESTAMP           DEFAULT current_timestamp,

    author_id   INTEGER             NOT NULL REFERENCES usr (id) ON DELETE CASCADE,
    forum_id    INTEGER             NOT NULL REFERENCES forum (id) ON DELETE CASCADE,
    CONSTRAINT unique_post UNIQUE (id, thread_id)
);

CREATE INDEX ON post USING HASH ((LOWER(forum)));
CREATE INDEX ON post (thread_id);

CREATE TABLE IF NOT EXISTS vote
(
    nickname    VARCHAR(128)        NOT NULL REFERENCES usr (nickname) ON DELETE CASCADE,
    voice       INTEGER             NOT NULL,

    thread_id   INTEGER             NOT NULL REFERENCES thread (id) ON DELETE CASCADE,
    usr_id      INTEGER             NOT NULL REFERENCES usr (id) ON DELETE CASCADE,
    CONSTRAINT unique_vote UNIQUE (usr_id, thread_id)
);

CREATE INDEX ON vote (usr_id);


-- Table that stores max id from every table
CREATE TABLE IF NOT EXISTS summary
(
    users       INTEGER             NOT NULL DEFAULT 0,
    forums      INTEGER             NOT NULL DEFAULT 0,
    threads     INTEGER             NOT NULL DEFAULT 0,
    posts       INTEGER             NOT NULL DEFAULT 0
);

INSERT INTO summary DEFAULT VALUES;

-- Change isEdited field on post on update
DROP FUNCTION IF EXISTS trigger_update_post();
CREATE FUNCTION trigger_update_post()
    RETURNS trigger AS
    $BODY$
    BEGIN
        IF NEW.message <> OLD.message THEN
            NEW.isedited = true;
        END IF;
        RETURN NEW;
    END;
    $BODY$
    LANGUAGE plpgsql;
CREATE TRIGGER update_post
    BEFORE UPDATE ON post
    FOR EACH ROW
    EXECUTE PROCEDURE trigger_update_post();


-- Increment threads on forum
DROP FUNCTION IF EXISTS trigger_increment_threads();
CREATE FUNCTION trigger_increment_threads()
RETURNS trigger AS
    $BODY$
    BEGIN
        UPDATE forum
            SET threads = threads + 1
        WHERE id = NEW.forum_id;
        RETURN NEW;
    END;
    $BODY$ LANGUAGE plpgsql;
CREATE TRIGGER increment_threads
    AFTER INSERT ON thread
    FOR EACH ROW
    EXECUTE PROCEDURE trigger_increment_threads();


-- Increment posts on forum
-- TODO: inserting pack of posts so remove this maybe
DROP FUNCTION IF EXISTS trigger_increment_posts();
CREATE FUNCTION trigger_increment_posts()
RETURNS trigger AS
    $BODY$
    BEGIN
        UPDATE forum
            SET posts = posts + 1
        WHERE id = NEW.forum_id;
        RETURN NEW;
    END;
    $BODY$ LANGUAGE plpgsql;
CREATE TRIGGER increment_posts
    AFTER INSERT ON post
    FOR EACH ROW
    EXECUTE PROCEDURE trigger_increment_posts();



COMMIT;
