-- +goose up
CREATE TABLE users
(
    id                bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email             text        NOT NULL,
    password_hash     text        NOT NULL,
    timezone          text        NOT NULL DEFAULT 'UTC',
    ui_language       text        NOT NULL DEFAULT 'ru' CHECK (ui_language IN ('ru', 'en')),
    english_variant   text        NOT NULL DEFAULT 'us' CHECK (english_variant IN ('us', 'uk')),
    desired_retention real        NOT NULL DEFAULT 0.9 CHECK (desired_retention BETWEEN 0.7 AND 0.99),
    fsrs_params       double precision[], -- NULL = веса по умолчанию
    created_at        timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX users_email_key ON users (lower(email));

-- Вход по почте и паролю
CREATE TABLE sessions
(
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    bigint      NOT NULL REFERENCES users ON DELETE CASCADE,
    token_hash text        NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX sessions_user_id_idx ON sessions (user_id);

-- Токены для MCP
CREATE TABLE api_tokens
(
    id           bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id      bigint      NOT NULL REFERENCES users ON DELETE CASCADE,
    name         text        NOT NULL,
    token_hash   text        NOT NULL UNIQUE,
    last_used_at timestamptz,
    created_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX api_tokens_user_id_idx ON api_tokens (user_id);

CREATE TABLE decks
(
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    owner_id    bigint REFERENCES users ON DELETE CASCADE, -- NULL = готовая системная
    title       text        NOT NULL,
    lang_learn  text        NOT NULL DEFAULT 'en',
    lang_native text        NOT NULL DEFAULT 'ru',
    is_public   boolean     NOT NULL DEFAULT false,
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX decks_owner_id_idx ON decks (owner_id);

-- «Я учу эту колоду»; у личной колоды тоже есть строка
CREATE TABLE enrollments
(
    user_id     bigint      NOT NULL REFERENCES users ON DELETE CASCADE,
    deck_id     bigint      NOT NULL REFERENCES decks ON DELETE CASCADE,
    deadline    date, -- до даты всё показывается ежедневно
    new_per_day integer     NOT NULL DEFAULT 20 CHECK (new_per_day BETWEEN 0 AND 200),
    joined_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, deck_id)
);
CREATE INDEX enrollments_deck_id_idx ON enrollments (deck_id);

CREATE TABLE items
(
    id                  bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    deck_id             bigint      NOT NULL REFERENCES decks ON DELETE CASCADE,
    kind                text        NOT NULL CHECK (kind IN ('word', 'phrase', 'sentence')),
    text                text        NOT NULL,
    translation         text        NOT NULL,
    ipa                 text,
    audio_path          text,        -- файл на диске VDS
    image_path          text,
    example             text,
    example_translation text,
    created_via         text        NOT NULL DEFAULT 'manual' CHECK (created_via IN ('manual', 'mcp', 'import')),
    archived_at         timestamptz, -- вместо удаления: история ответов остаётся
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);
-- Защита от дублей при добавлении через MCP
CREATE UNIQUE INDEX items_deck_text_key ON items (deck_id, lower(text)) WHERE archived_at IS NULL;

-- Строка на пользователя, запись и навык
CREATE TABLE progress
(
    user_id          bigint  NOT NULL REFERENCES users ON DELETE CASCADE,
    item_id          bigint  NOT NULL REFERENCES items ON DELETE CASCADE,
    skill            text    NOT NULL CHECK (skill IN ('recognize', 'recall', 'listen', 'use')),
    state            text    NOT NULL DEFAULT 'new' CHECK (state IN ('new', 'learning', 'review', 'relearning')),
    stability        double precision, -- FSRS, в днях
    difficulty       double precision, -- FSRS, 1–10
    due_at           timestamptz,
    last_reviewed_at timestamptz,
    lapses           integer NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, item_id, skill)
);
CREATE INDEX progress_due_idx ON progress (user_id, due_at) WHERE due_at IS NOT NULL;
CREATE INDEX progress_item_id_idx ON progress (item_id);

-- История ответов: только добавляется, по ней пересчитывается progress
CREATE TABLE reviews
(
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id     bigint      NOT NULL REFERENCES users ON DELETE CASCADE,
    item_id     bigint      NOT NULL REFERENCES items ON DELETE RESTRICT,
    skill       text        NOT NULL CHECK (skill IN ('recognize', 'recall', 'listen', 'use')),
    exercise    text        NOT NULL CHECK (exercise IN ('choice', 'typing', 'cloze', 'dictation', 'compose')),
    rating      smallint    NOT NULL CHECK (rating BETWEEN 1 AND 4), -- 1 Again … 4 Easy
    answer      text,
    hints_used  smallint    NOT NULL DEFAULT 0,
    response_ms integer,
    feedback    text,                                                -- разбор от ИИ
    reviewed_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX reviews_user_time_idx ON reviews (user_id, reviewed_at);
CREATE INDEX reviews_progress_idx ON reviews (user_id, item_id, skill, reviewed_at);

-- +goose down
DROP TABLE reviews;
DROP TABLE progress;
DROP TABLE items;
DROP TABLE enrollments;
DROP TABLE decks;
DROP TABLE api_tokens;
DROP TABLE sessions;
DROP TABLE users;
