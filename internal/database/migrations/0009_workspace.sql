CREATE TABLE app_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO app_settings(key,value) VALUES
    ('density','compact'),
    ('sidebar_collapsed','false'),
    ('ai_provider','ollama'),
    ('ai_endpoint','http://127.0.0.1:11434'),
    ('ai_model','qwen2.5:7b'),
    ('ai_mode','strict'),
    ('reduce_motion','false');

ALTER TABLE accounts ADD COLUMN notes TEXT NOT NULL DEFAULT '';
ALTER TABLE accounts ADD COLUMN astrite INTEGER NOT NULL DEFAULT 0;
ALTER TABLE accounts ADD COLUMN radiant_tides INTEGER NOT NULL DEFAULT 0;
ALTER TABLE accounts ADD COLUMN active INTEGER NOT NULL DEFAULT 1 CHECK(active IN (0,1));

CREATE TABLE planner_goals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    goal_type TEXT NOT NULL DEFAULT 'character',
    target_name TEXT NOT NULL DEFAULT '',
    required_amount INTEGER NOT NULL DEFAULT 0,
    owned_amount INTEGER NOT NULL DEFAULT 0,
    shell_credits INTEGER NOT NULL DEFAULT 0,
    priority INTEGER NOT NULL DEFAULT 2,
    due_date TEXT NOT NULL DEFAULT '',
    completed INTEGER NOT NULL DEFAULT 0 CHECK(completed IN (0,1)),
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE convene_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    banner TEXT NOT NULL,
    banner_type TEXT NOT NULL DEFAULT 'featured_character',
    item_name TEXT NOT NULL,
    rarity INTEGER NOT NULL CHECK(rarity BETWEEN 3 AND 5),
    pull_number INTEGER NOT NULL DEFAULT 1,
    guaranteed INTEGER NOT NULL DEFAULT 0 CHECK(guaranteed IN (0,1)),
    obtained_at TEXT NOT NULL,
    notes TEXT NOT NULL DEFAULT ''
);

CREATE TABLE enemies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    level INTEGER NOT NULL DEFAULT 90,
    resistance REAL NOT NULL DEFAULT 0.10,
    damage_reduction REAL NOT NULL DEFAULT 0,
    element_reduction REAL NOT NULL DEFAULT 0,
    notes TEXT NOT NULL DEFAULT ''
);
INSERT INTO enemies(name,level,resistance,notes) VALUES
    ('Alvo padrão',90,0.10,'Cenário neutro para theorycraft'),
    ('Inimigo resistente',100,0.40,'Cenário de resistência elevada');

CREATE TABLE formula_versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    game_version TEXT NOT NULL DEFAULT '',
    defense_constant REAL NOT NULL DEFAULT 800,
    level_factor REAL NOT NULL DEFAULT 8,
    confidence TEXT NOT NULL DEFAULT 'community_tested',
    references_text TEXT NOT NULL DEFAULT '',
    rounding_policy TEXT NOT NULL DEFAULT 'full_precision',
    active INTEGER NOT NULL DEFAULT 0 CHECK(active IN (0,1)),
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO formula_versions(name,confidence,references_text,active)
VALUES('WaveArchive 1.0','community_tested','Fórmula comunitária versionada; validar após mudanças do jogo.',1);

ALTER TABLE builds ADD COLUMN target_enemy_id INTEGER REFERENCES enemies(id);
ALTER TABLE builds ADD COLUMN rotation_id INTEGER REFERENCES rotations(id);
ALTER TABLE builds ADD COLUMN conditions TEXT NOT NULL DEFAULT '';

CREATE TABLE build_versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    build_id INTEGER NOT NULL,
    snapshot_json TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(build_id) REFERENCES builds(id) ON DELETE CASCADE
);
CREATE INDEX build_versions_build_idx ON build_versions(build_id,id DESC);
CREATE INDEX planner_goals_priority_idx ON planner_goals(completed,priority,id);
CREATE INDEX convene_records_date_idx ON convene_records(obtained_at DESC,id DESC);
