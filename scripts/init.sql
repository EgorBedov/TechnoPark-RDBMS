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
    title       VARCHAR(1024)       NOT NULL,
    usr         VARCHAR(128)        NOT NULL,
    slug        VARCHAR(128)        NOT NULL UNIQUE,
    posts       BIGINT              DEFAULT 0,
    threads     INTEGER             DEFAULT 0,

    id          SERIAL              PRIMARY KEY,
    usr_id      INTEGER             REFERENCES usr (id)
);

CREATE TABLE IF NOT EXISTS thread
(
    id          SERIAL              PRIMARY KEY,
    title       VARCHAR(1024)       NOT NULL,
    author      VARCHAR(128)        REFERENCES usr (nickname),
    forum       VARCHAR(128)        REFERENCES forum (slug),
    message     TEXT                DEFAULT NULL,
    votes       INTEGER             DEFAULT 0,
    slug        VARCHAR(128)        NOT NULL,
    created     TIMESTAMP           DEFAULT current_timestamp,

    author_id   INTEGER             REFERENCES usr (id),
    forum_id    INTEGER             REFERENCES forum (id)
);

CREATE TABLE IF NOT EXISTS post
(
    id          BIGSERIAL           PRIMARY KEY,
    parent      BIGINT              DEFAULT 0,
    author      VARCHAR(128)        REFERENCES usr (nickname),
    message     TEXT                NOT NULL,
    isEdited    BOOLEAN             DEFAULT FALSE,
    forum       VARCHAR(128)        REFERENCES forum (slug),
    thread_id   INTEGER             REFERENCES thread (id),
    created     TIMESTAMP           DEFAULT current_timestamp,

    author_id   INTEGER             REFERENCES usr (id),
    forum_id    INTEGER             REFERENCES forum (id)
);

CREATE TABLE IF NOT EXISTS vote
(
    nickname    VARCHAR(128)        REFERENCES usr (nickname),
    voice       INTEGER             NOT NULL,

    usr_id      INTEGER             REFERENCES usr (id)
);

COMMIT;
