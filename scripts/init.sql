BEGIN;

CREATE TABLE IF NOT EXISTS usr
(
    nickname    VARCHAR(128)        NOT NULL UNIQUE,
    fullname    VARCHAR(256)        NOT NULL,
    about       TEXT                DEFAULT NULL,
    email       VARCHAR(128)        UNIQUE,

    id          SERIAL              PRIMARY KEY
);

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

CREATE TABLE IF NOT EXISTS thread
(
    id          SERIAL              PRIMARY KEY,
    title       VARCHAR(1024)       NOT NULL,
    author      VARCHAR(128)        REFERENCES usr (nickname) ON DELETE CASCADE,
    forum       VARCHAR(128)        REFERENCES forum (slug) ON DELETE CASCADE,
    message     TEXT                DEFAULT NULL,
    votes       INTEGER             DEFAULT 0,
    slug        VARCHAR(128)        NOT NULL UNIQUE,
    created     TIMESTAMP           DEFAULT current_timestamp,

    author_id   INTEGER             NOT NULL REFERENCES usr (id) ON DELETE CASCADE,
    forum_id    INTEGER             NOT NULL REFERENCES forum (id) ON DELETE CASCADE
);

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
    forum_id    INTEGER             NOT NULL REFERENCES forum (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS vote
(
    nickname    VARCHAR(128)        NOT NULL REFERENCES usr (nickname) ON DELETE CASCADE,
    voice       INTEGER             NOT NULL,

    thread_id   INTEGER             NOT NULL REFERENCES thread (id) ON DELETE CASCADE,
    usr_id      INTEGER             NOT NULL REFERENCES usr (id) ON DELETE CASCADE,
    CONSTRAINT unique_vote UNIQUE (usr_id, thread_id)
);

CREATE TABLE IF NOT EXISTS summary
(
    users       INTEGER             NOT NULL DEFAULT 0,
    forums      INTEGER             NOT NULL DEFAULT 0,
    threads     INTEGER             NOT NULL DEFAULT 0,
    posts       INTEGER             NOT NULL DEFAULT 0
);

INSERT INTO summary DEFAULT VALUES;

COMMIT;
