CREATE TABLE builds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    character_id INTEGER NOT NULL,
    character_level INTEGER NOT NULL DEFAULT 90 CHECK (character_level BETWEEN 1 AND 90),
    sequence INTEGER NOT NULL DEFAULT 0 CHECK (sequence BETWEEN 0 AND 6),
    weapon_id INTEGER,
    weapon_level INTEGER NOT NULL DEFAULT 90 CHECK (weapon_level BETWEEN 1 AND 90),
    weapon_rank INTEGER NOT NULL DEFAULT 1 CHECK (weapon_rank BETWEEN 1 AND 5),
    notes TEXT NOT NULL DEFAULT '',
    favorite INTEGER NOT NULL DEFAULT 0 CHECK (favorite IN (0, 1)),
    locked INTEGER NOT NULL DEFAULT 0 CHECK (locked IN (0, 1)),
    game_version TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TEXT,
    FOREIGN KEY (character_id) REFERENCES characters(id),
    FOREIGN KEY (weapon_id) REFERENCES weapons(id)
);

CREATE INDEX builds_character_idx ON builds(character_id);
CREATE INDEX builds_updated_idx ON builds(updated_at DESC);
CREATE INDEX builds_deleted_idx ON builds(deleted_at);
