ALTER TABLE characters ADD COLUMN description TEXT NOT NULL DEFAULT '';
ALTER TABLE characters ADD COLUMN birthday TEXT NOT NULL DEFAULT '';
ALTER TABLE characters ADD COLUMN gender TEXT NOT NULL DEFAULT '';
ALTER TABLE characters ADD COLUMN region TEXT NOT NULL DEFAULT '';
ALTER TABLE characters ADD COLUMN faction TEXT NOT NULL DEFAULT '';
ALTER TABLE characters ADD COLUMN talent_name TEXT NOT NULL DEFAULT '';
ALTER TABLE characters ADD COLUMN talent_description TEXT NOT NULL DEFAULT '';
ALTER TABLE characters ADD COLUMN signature_weapon_id INTEGER;
ALTER TABLE characters ADD COLUMN detail_loaded INTEGER NOT NULL DEFAULT 0;

CREATE TABLE weapons (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    rarity INTEGER NOT NULL,
    weapon_type INTEGER NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    effect_name TEXT NOT NULL DEFAULT '',
    effect TEXT NOT NULL DEFAULT '',
    icon_path TEXT NOT NULL DEFAULT '',
    params_json TEXT NOT NULL DEFAULT '[]',
    game_version TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE skills (
    character_id INTEGER NOT NULL,
    node_id TEXT NOT NULL,
    skill_type TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    icon_path TEXT NOT NULL DEFAULT '',
    levels_json TEXT NOT NULL DEFAULT '{}',
    sort_order INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (character_id, node_id),
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
);

CREATE TABLE resonance_chains (
    character_id INTEGER NOT NULL,
    sequence INTEGER NOT NULL CHECK (sequence BETWEEN 1 AND 6),
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    icon_path TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (character_id, sequence),
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
);

CREATE INDEX skills_character_order_idx ON skills(character_id, sort_order);
CREATE INDEX chains_character_sequence_idx ON resonance_chains(character_id, sequence);
CREATE INDEX characters_signature_weapon_idx ON characters(signature_weapon_id);
