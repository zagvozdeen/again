-- +goose up
CREATE TYPE ui_language AS ENUM ('ru', 'en');
CREATE TYPE item_kind AS ENUM ('word', 'phrase', 'sentence');
CREATE TYPE item_source AS ENUM ('manual', 'mcp', 'import');
CREATE TYPE skill AS ENUM ('recognize', 'recall', 'listen', 'use');
CREATE TYPE progress_state AS ENUM ('new', 'learning', 'review', 'relearning');
CREATE TYPE exercise AS ENUM ('choice', 'typing', 'cloze', 'dictation', 'compose');

CREATE TABLE IF NOT EXISTS users
(
    id                SERIAL PRIMARY KEY,
    name              VARCHAR(255)       NOT NULL,
    email             VARCHAR(255)       NOT NULL UNIQUE,
    password_hash     VARCHAR(255)       NOT NULL,
    ui_language       ui_language        NOT NULL,
    desired_retention REAL               NOT NULL,
    fsrs_params       DOUBLE PRECISION[] NULL,
    created_at        TIMESTAMPTZ        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMPTZ        NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions
(
    id         SERIAL PRIMARY KEY,
    user_id    INTEGER      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS api_tokens
(
    id           SERIAL PRIMARY KEY,
    user_id      INTEGER      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    name         VARCHAR(255) NOT NULL,
    token_hash   VARCHAR(255) NOT NULL UNIQUE,
    last_used_at TIMESTAMPTZ  NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS decks
(
    id          SERIAL PRIMARY KEY,
    owner_id    INTEGER      NULL REFERENCES users (id) ON DELETE RESTRICT,
    title       VARCHAR(255) NOT NULL,
    lang_learn  VARCHAR(255) NOT NULL,
    lang_native VARCHAR(255) NOT NULL,
    is_public   BOOLEAN      NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS enrollments
(
    user_id     INTEGER     NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    deck_id     INTEGER     NOT NULL REFERENCES decks (id) ON DELETE RESTRICT,
    deadline    TIMESTAMPTZ NULL,
    new_per_day INTEGER     NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, deck_id)
);

CREATE TABLE IF NOT EXISTS items
(
    id                  SERIAL PRIMARY KEY,
    deck_id             INTEGER      NOT NULL REFERENCES decks (id) ON DELETE RESTRICT,
    kind                item_kind    NOT NULL,
    text                VARCHAR(255) NOT NULL,
    translation         VARCHAR(255) NOT NULL,
    ipa                 VARCHAR(255) NULL,
    audio_path          VARCHAR(255) NULL,
    image_path          VARCHAR(255) NULL,
    example             VARCHAR(255) NULL,
    example_translation VARCHAR(255) NULL,
    created_via         item_source  NOT NULL,
    archived_at         TIMESTAMPTZ  NULL,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS progress
(
    user_id          INTEGER          NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    item_id          INTEGER          NOT NULL REFERENCES items (id) ON DELETE RESTRICT,
    skill            skill            NOT NULL,
    state            progress_state   NOT NULL,
    stability        DOUBLE PRECISION NULL,
    difficulty       DOUBLE PRECISION NULL,
    due_at           TIMESTAMPTZ      NULL,
    last_reviewed_at TIMESTAMPTZ      NULL,
    lapses           INTEGER          NOT NULL,
    created_at       TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, item_id, skill)
);

CREATE TABLE IF NOT EXISTS reviews
(
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    item_id     INTEGER      NOT NULL REFERENCES items (id) ON DELETE RESTRICT,
    skill       skill        NOT NULL,
    exercise    exercise     NOT NULL,
    rating      SMALLINT     NOT NULL,
    answer      VARCHAR(255) NULL,
    hints_used  SMALLINT     NOT NULL,
    response_ms INTEGER      NULL,
    feedback    TEXT         NULL,
    reviewed_at TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose down
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS progress;
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS enrollments;
DROP TABLE IF EXISTS decks;
DROP TABLE IF EXISTS api_tokens;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS exercise;
DROP TYPE IF EXISTS progress_state;
DROP TYPE IF EXISTS skill;
DROP TYPE IF EXISTS item_source;
DROP TYPE IF EXISTS item_kind;
DROP TYPE IF EXISTS ui_language;
