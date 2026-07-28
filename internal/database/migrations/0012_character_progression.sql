CREATE TABLE materials (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    rarity INTEGER NOT NULL DEFAULT 0,
    material_type INTEGER NOT NULL DEFAULT 0,
    description TEXT NOT NULL DEFAULT '',
    icon_path TEXT NOT NULL DEFAULT '',
    sources_json TEXT NOT NULL DEFAULT '[]',
    game_version TEXT NOT NULL
);

CREATE TABLE character_ascension_costs (
    character_id INTEGER NOT NULL,
    stage INTEGER NOT NULL,
    material_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    PRIMARY KEY (character_id, stage, material_id),
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE,
    FOREIGN KEY (material_id) REFERENCES materials(id)
);

CREATE TABLE character_level_exp (
    character_id INTEGER NOT NULL,
    level INTEGER NOT NULL,
    experience INTEGER NOT NULL,
    PRIMARY KEY (character_id, level),
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
);

CREATE TABLE character_stats (
    character_id INTEGER NOT NULL,
    ascension INTEGER NOT NULL,
    level INTEGER NOT NULL,
    hp REAL NOT NULL,
    atk REAL NOT NULL,
    def REAL NOT NULL,
    PRIMARY KEY (character_id, ascension, level),
    FOREIGN KEY (character_id) REFERENCES characters(id) ON DELETE CASCADE
);

CREATE TABLE skill_progression (
    character_id INTEGER NOT NULL,
    node_id TEXT NOT NULL,
    node_type INTEGER NOT NULL DEFAULT 0,
    max_level INTEGER NOT NULL DEFAULT 1,
    values_json TEXT NOT NULL DEFAULT '[]',
    PRIMARY KEY (character_id, node_id),
    FOREIGN KEY (character_id, node_id) REFERENCES skills(character_id, node_id) ON DELETE CASCADE
);

CREATE TABLE skill_unlock_costs (
    character_id INTEGER NOT NULL,
    node_id TEXT NOT NULL,
    material_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    PRIMARY KEY (character_id, node_id, material_id),
    FOREIGN KEY (character_id, node_id) REFERENCES skills(character_id, node_id) ON DELETE CASCADE,
    FOREIGN KEY (material_id) REFERENCES materials(id)
);

CREATE TABLE skill_level_costs (
    character_id INTEGER NOT NULL,
    node_id TEXT NOT NULL,
    level INTEGER NOT NULL,
    material_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL,
    PRIMARY KEY (character_id, node_id, level, material_id),
    FOREIGN KEY (character_id, node_id) REFERENCES skills(character_id, node_id) ON DELETE CASCADE,
    FOREIGN KEY (material_id) REFERENCES materials(id)
);

CREATE INDEX character_ascension_stage_idx ON character_ascension_costs(character_id, stage);
CREATE INDEX skill_level_cost_idx ON skill_level_costs(character_id, node_id, level);
