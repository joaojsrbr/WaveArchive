CREATE TABLE teams (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    notes TEXT NOT NULL DEFAULT '',
    favorite INTEGER NOT NULL DEFAULT 0 CHECK (favorite IN (0,1)),
    locked INTEGER NOT NULL DEFAULT 0 CHECK (locked IN (0,1)),
    game_version TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TEXT
);

CREATE TABLE team_members (
    team_id INTEGER NOT NULL,
    slot INTEGER NOT NULL CHECK (slot BETWEEN 1 AND 3),
    character_id INTEGER NOT NULL,
    build_id INTEGER,
    role TEXT NOT NULL DEFAULT '',
    custom_role TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (team_id, slot),
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    FOREIGN KEY (character_id) REFERENCES characters(id),
    FOREIGN KEY (build_id) REFERENCES builds(id)
);

CREATE INDEX teams_updated_idx ON teams(updated_at DESC);
CREATE INDEX teams_deleted_idx ON teams(deleted_at);
CREATE INDEX team_members_character_idx ON team_members(character_id);
