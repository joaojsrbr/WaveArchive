CREATE TABLE IF NOT EXISTS game_versions (
    version TEXT PRIMARY KEY,
    synced_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS characters (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    nickname TEXT NOT NULL DEFAULT '',
    rarity INTEGER NOT NULL CHECK (rarity BETWEEN 0 AND 5),
    element INTEGER NOT NULL DEFAULT 0,
    weapon_type INTEGER NOT NULL DEFAULT 0,
    icon_path TEXT NOT NULL DEFAULT '',
    background_path TEXT NOT NULL DEFAULT '',
    game_version TEXT NOT NULL,
    source_payload TEXT,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS owned_characters (
    profile_id INTEGER NOT NULL DEFAULT 1,
    character_id INTEGER NOT NULL,
    level INTEGER NOT NULL DEFAULT 1,
    sequence INTEGER NOT NULL DEFAULT 0 CHECK (sequence BETWEEN 0 AND 6),
    favorite INTEGER NOT NULL DEFAULT 0 CHECK (favorite IN (0, 1)),
    PRIMARY KEY (profile_id, character_id),
    FOREIGN KEY (character_id) REFERENCES characters(id)
);

CREATE VIRTUAL TABLE IF NOT EXISTS characters_fts USING fts5(
    name,
    nickname,
    content='characters',
    content_rowid='id',
    tokenize='unicode61 remove_diacritics 2'
);

CREATE TRIGGER IF NOT EXISTS characters_ai AFTER INSERT ON characters BEGIN
    INSERT INTO characters_fts(rowid, name, nickname)
    VALUES (new.id, new.name, new.nickname);
END;

CREATE TRIGGER IF NOT EXISTS characters_ad AFTER DELETE ON characters BEGIN
    INSERT INTO characters_fts(characters_fts, rowid, name, nickname)
    VALUES ('delete', old.id, old.name, old.nickname);
END;

CREATE TRIGGER IF NOT EXISTS characters_au AFTER UPDATE ON characters BEGIN
    INSERT INTO characters_fts(characters_fts, rowid, name, nickname)
    VALUES ('delete', old.id, old.name, old.nickname);
    INSERT INTO characters_fts(rowid, name, nickname)
    VALUES (new.id, new.name, new.nickname);
END;

CREATE INDEX IF NOT EXISTS characters_element_idx ON characters(element);
CREATE INDEX IF NOT EXISTS characters_rarity_idx ON characters(rarity);
