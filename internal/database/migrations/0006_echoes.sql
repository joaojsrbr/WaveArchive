CREATE TABLE echo_sets (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    icon_path TEXT NOT NULL DEFAULT '',
    two_piece TEXT NOT NULL DEFAULT '',
    five_piece TEXT NOT NULL DEFAULT '',
    game_version TEXT NOT NULL
);

CREATE TABLE echoes (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    code TEXT NOT NULL DEFAULT '',
    echo_type TEXT NOT NULL DEFAULT '',
    class TEXT NOT NULL DEFAULT '',
    cost INTEGER NOT NULL DEFAULT 1,
    place TEXT NOT NULL DEFAULT '',
    icon_path TEXT NOT NULL DEFAULT '',
    skill TEXT NOT NULL DEFAULT '',
    rarities_json TEXT NOT NULL DEFAULT '[]',
    sonata_ids_json TEXT NOT NULL DEFAULT '[]',
    game_version TEXT NOT NULL
);

CREATE TABLE owned_echoes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    profile_id INTEGER NOT NULL DEFAULT 1,
    echo_id INTEGER NOT NULL,
    main_stat TEXT NOT NULL DEFAULT '',
    substats_json TEXT NOT NULL DEFAULT '[]',
    level INTEGER NOT NULL DEFAULT 0 CHECK (level BETWEEN 0 AND 25),
    sonata_id INTEGER,
    character_id INTEGER,
    locked INTEGER NOT NULL DEFAULT 0 CHECK (locked IN (0,1)),
    favorite INTEGER NOT NULL DEFAULT 0 CHECK (favorite IN (0,1)),
    note TEXT NOT NULL DEFAULT '',
    FOREIGN KEY (echo_id) REFERENCES echoes(id),
    FOREIGN KEY (sonata_id) REFERENCES echo_sets(id),
    FOREIGN KEY (character_id) REFERENCES characters(id)
);

CREATE TABLE build_echoes (
    build_id INTEGER NOT NULL,
    slot INTEGER NOT NULL CHECK (slot BETWEEN 1 AND 5),
    owned_echo_id INTEGER NOT NULL,
    PRIMARY KEY (build_id, slot),
    FOREIGN KEY (build_id) REFERENCES builds(id) ON DELETE CASCADE,
    FOREIGN KEY (owned_echo_id) REFERENCES owned_echoes(id)
);

CREATE VIRTUAL TABLE echoes_fts USING fts5(name, skill, content='echoes', content_rowid='id');
CREATE TRIGGER echoes_ai AFTER INSERT ON echoes BEGIN
  INSERT INTO echoes_fts(rowid,name,skill) VALUES(new.id,new.name,new.skill);
END;
CREATE TRIGGER echoes_au AFTER UPDATE ON echoes BEGIN
  INSERT INTO echoes_fts(echoes_fts,rowid,name,skill) VALUES('delete',old.id,old.name,old.skill);
  INSERT INTO echoes_fts(rowid,name,skill) VALUES(new.id,new.name,new.skill);
END;
CREATE TRIGGER echoes_ad AFTER DELETE ON echoes BEGIN
  INSERT INTO echoes_fts(echoes_fts,rowid,name,skill) VALUES('delete',old.id,old.name,old.skill);
END;
CREATE INDEX echoes_cost_idx ON echoes(cost);
CREATE INDEX owned_echoes_echo_idx ON owned_echoes(echo_id);
