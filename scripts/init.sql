BEGIN;

CREATE EXTENSION IF NOT EXISTS citext;

CREATE UNLOGGED TABLE IF NOT EXISTS usr
(
    nickname    citext              PRIMARY KEY NOT NULL UNIQUE,
    fullname    VARCHAR(256)        NOT NULL,
    about       TEXT                DEFAULT NULL,
    email       citext              UNIQUE
);

CREATE INDEX ON usr (nickname DESC);

CREATE UNLOGGED TABLE IF NOT EXISTS forum
(
    title       VARCHAR(1024)       NOT NULL,
    usr         citext              NOT NULL REFERENCES usr (nickname) ON DELETE CASCADE,
    slug        citext              PRIMARY KEY NOT NULL UNIQUE,
    posts       BIGINT              DEFAULT 0,
    threads     INTEGER             DEFAULT 0
);

CREATE INDEX ON forum (usr);

CREATE UNLOGGED TABLE IF NOT EXISTS thread
(
    id          SERIAL              PRIMARY KEY,
    title       VARCHAR(1024)       NOT NULL,
    author      citext              REFERENCES usr (nickname) ON DELETE CASCADE,
    forum       citext              REFERENCES forum (slug) ON DELETE CASCADE,
    message     TEXT                DEFAULT NULL,
    votes       INTEGER             DEFAULT 0,
    slug        citext              UNIQUE,
    created     TIMESTAMP           DEFAULT current_timestamp
);


CREATE INDEX ON thread (forum, created);
CREATE INDEX ON thread (forum, author);

CREATE UNLOGGED TABLE IF NOT EXISTS post
(
    id          SERIAL              NOT NULL PRIMARY KEY,
    parent      INTEGER             DEFAULT 0,
    author      citext              REFERENCES usr (nickname) ON DELETE CASCADE,
    message     TEXT                NOT NULL,
    isEdited    BOOLEAN             DEFAULT FALSE,
    forum       citext              REFERENCES forum (slug) ON DELETE CASCADE,
    thread_id   INTEGER             REFERENCES thread (id) ON DELETE CASCADE,
    created     TIMESTAMP           DEFAULT current_timestamp,
    path        INTEGER[]           NOT NULL DEFAULT ARRAY[0]
);

CREATE INDEX ON post USING HASH (forum);
CREATE INDEX ON post (thread_id ASC);
CREATE INDEX ON post (thread_id, id ASC, path ASC) WHERE thread_id < 5000;
CREATE INDEX ON post (thread_id, id ASC, path ASC) WHERE thread_id >= 5000;
CREATE INDEX ON post (forum, author ASC);
CREATE INDEX ON post (thread_id ASC) WHERE parent = 0;
CREATE INDEX ON post (thread_id ASC, (path[1]) ASC) WHERE parent = 0;
CREATE INDEX ON post ((path[1]) ASC) WHERE parent = 0;
CREATE INDEX ON post (path ASC);
CREATE INDEX ON post ((path[1]) ASC);
CREATE INDEX ON post (id ASC, (path[1]) ASC);

-- Stores authors of posts and threads in a forum
CREATE UNLOGGED TABLE IF NOT EXISTS forum_authors (
    forum       citext          NOT NULL,
    author      citext          NOT NULL,
    CONSTRAINT unique_author UNIQUE (forum, author)
);
CREATE INDEX ON forum_authors (forum, author ASC);
CREATE INDEX ON forum_authors (forum, author DESC);


CREATE UNLOGGED TABLE IF NOT EXISTS vote
(
    nickname    citext      NOT NULL REFERENCES usr (nickname) ON DELETE CASCADE,
    voice       INTEGER     NOT NULL,
    thread_id   INTEGER     NOT NULL REFERENCES thread (id) ON DELETE CASCADE,
    CONSTRAINT unique_vote UNIQUE (nickname, thread_id)
);


-- Table that stores max id from every table
CREATE UNLOGGED TABLE IF NOT EXISTS summary
(
    users       INTEGER             NOT NULL DEFAULT 0,
    forums      INTEGER             NOT NULL DEFAULT 0,
    threads     INTEGER             NOT NULL DEFAULT 0,
    posts       INTEGER             NOT NULL DEFAULT 0
);

INSERT INTO summary DEFAULT VALUES;

-- Change isEdited field on post on update
CREATE OR REPLACE FUNCTION trigger_update_post()
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
CREATE OR REPLACE FUNCTION trigger_increment_threads()
RETURNS trigger AS
    $BODY$
    BEGIN
        UPDATE
            forum
        SET
            threads = threads + 1
        WHERE
            slug = NEW.forum;
        RETURN NEW;
    END;
    $BODY$ LANGUAGE plpgsql;
CREATE TRIGGER increment_threads
    AFTER INSERT ON thread
    FOR EACH ROW
    EXECUTE PROCEDURE trigger_increment_threads();



COMMIT;
