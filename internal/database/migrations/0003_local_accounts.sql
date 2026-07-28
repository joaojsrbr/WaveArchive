CREATE TABLE accounts (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO accounts(id, name) VALUES (1, 'Conta principal');

ALTER TABLE owned_characters
    ADD COLUMN owned INTEGER NOT NULL DEFAULT 1 CHECK (owned IN (0, 1));

CREATE INDEX owned_characters_owned_idx
    ON owned_characters(profile_id, owned);

CREATE INDEX owned_characters_favorite_idx
    ON owned_characters(profile_id, favorite);
