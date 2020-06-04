BEGIN;

CREATE EXTENSION IF NOT EXISTS citext;

CREATE UNLOGGED TABLE IF NOT EXISTS usr
(
    nickname    citext              PRIMARY KEY NOT NULL UNIQUE,
    fullname    VARCHAR(256)        NOT NULL,
    about       TEXT                DEFAULT NULL,
    email       citext              UNIQUE
);

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
    id          BIGSERIAL           PRIMARY KEY,
    parent      BIGINT              DEFAULT 0,
    author      citext              REFERENCES usr (nickname) ON DELETE CASCADE,
    message     TEXT                NOT NULL,
    isEdited    BOOLEAN             DEFAULT FALSE,
    forum       citext              REFERENCES forum (slug) ON DELETE CASCADE,
    thread_id   INTEGER             REFERENCES thread (id) ON DELETE CASCADE,
    created     TIMESTAMP           DEFAULT current_timestamp,
    root        INTEGER             DEFAULT 0
);

CREATE INDEX ON post USING HASH (forum);
CREATE INDEX ON post (parent);
CREATE INDEX ON post (thread_id);
CREATE INDEX ON post (root);
CREATE INDEX ON post (forum, author);
CREATE INDEX ON post (thread_id)
WHERE parent = 0;


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
