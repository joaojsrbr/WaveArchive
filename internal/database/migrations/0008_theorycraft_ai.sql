CREATE TABLE build_configs (
    build_id INTEGER PRIMARY KEY,
    scaling_type TEXT NOT NULL DEFAULT 'ATK',
    base_atk REAL NOT NULL DEFAULT 1000,
    base_hp REAL NOT NULL DEFAULT 10000,
    base_def REAL NOT NULL DEFAULT 1000,
    motion_value REAL NOT NULL DEFAULT 2,
    flat_damage REAL NOT NULL DEFAULT 0,
    enemy_level INTEGER NOT NULL DEFAULT 90,
    enemy_resistance REAL NOT NULL DEFAULT 0.1,
    defense_ignore REAL NOT NULL DEFAULT 0,
    damage_reduction REAL NOT NULL DEFAULT 0,
    element_reduction REAL NOT NULL DEFAULT 0,
    extra_damage_bonuses_json TEXT NOT NULL DEFAULT '[]',
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (build_id) REFERENCES builds(id) ON DELETE CASCADE
);

CREATE TABLE buffs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    team_id INTEGER NOT NULL,
    source_slot INTEGER NOT NULL CHECK (source_slot BETWEEN 1 AND 3),
    target_slot INTEGER NOT NULL DEFAULT 0 CHECK (target_slot BETWEEN 0 AND 3),
    name TEXT NOT NULL,
    modifier_group TEXT NOT NULL,
    value REAL NOT NULL,
    scope TEXT NOT NULL DEFAULT 'TEAM',
    condition_text TEXT NOT NULL DEFAULT '',
    assume_active INTEGER NOT NULL DEFAULT 1 CHECK (assume_active IN (0,1)),
    duration REAL NOT NULL DEFAULT 0,
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE
);

CREATE TABLE rotations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    team_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    duration REAL NOT NULL DEFAULT 0,
    notes TEXT NOT NULL DEFAULT '',
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE
);

CREATE TABLE rotation_actions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rotation_id INTEGER NOT NULL,
    sort_order INTEGER NOT NULL,
    slot INTEGER NOT NULL CHECK (slot BETWEEN 1 AND 3),
    action_type TEXT NOT NULL DEFAULT 'SKILL',
    name TEXT NOT NULL,
    motion_value REAL NOT NULL DEFAULT 0,
    cast_time REAL NOT NULL DEFAULT 0,
    energy REAL NOT NULL DEFAULT 0,
    concerto REAL NOT NULL DEFAULT 0,
    FOREIGN KEY (rotation_id) REFERENCES rotations(id) ON DELETE CASCADE
);

CREATE TABLE ai_conversations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    context_type TEXT NOT NULL DEFAULT 'general',
    context_id INTEGER,
    provider TEXT NOT NULL DEFAULT 'ollama',
    model TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE ai_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id INTEGER NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('user','assistant')),
    content TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conversation_id) REFERENCES ai_conversations(id) ON DELETE CASCADE
);

CREATE INDEX buffs_team_idx ON buffs(team_id);
CREATE INDEX rotations_team_idx ON rotations(team_id);
CREATE INDEX rotation_actions_order_idx ON rotation_actions(rotation_id,sort_order);
CREATE INDEX ai_messages_conversation_idx ON ai_messages(conversation_id,id);
