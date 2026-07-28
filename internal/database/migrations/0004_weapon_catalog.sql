ALTER TABLE weapons ADD COLUMN base_atk INTEGER NOT NULL DEFAULT 0;
ALTER TABLE weapons ADD COLUMN sub_stat TEXT NOT NULL DEFAULT '';

CREATE TABLE owned_weapons (
    profile_id INTEGER NOT NULL DEFAULT 1,
    weapon_id INTEGER NOT NULL,
    level INTEGER NOT NULL DEFAULT 1 CHECK (level BETWEEN 1 AND 90),
    weapon_rank INTEGER NOT NULL DEFAULT 1 CHECK (weapon_rank BETWEEN 1 AND 5),
    favorite INTEGER NOT NULL DEFAULT 0 CHECK (favorite IN (0, 1)),
    owned INTEGER NOT NULL DEFAULT 1 CHECK (owned IN (0, 1)),
    PRIMARY KEY (profile_id, weapon_id),
    FOREIGN KEY (weapon_id) REFERENCES weapons(id)
);

CREATE VIRTUAL TABLE weapons_fts USING fts5(
    name,
    description,
    effect_name,
    content='weapons',
    content_rowid='id',
    tokenize='unicode61 remove_diacritics 2'
);

CREATE TRIGGER weapons_ai AFTER INSERT ON weapons BEGIN
    INSERT INTO weapons_fts(rowid, name, description, effect_name)
    VALUES (new.id, new.name, new.description, new.effect_name);
END;

CREATE TRIGGER weapons_ad AFTER DELETE ON weapons BEGIN
    INSERT INTO weapons_fts(weapons_fts, rowid, name, description, effect_name)
    VALUES ('delete', old.id, old.name, old.description, old.effect_name);
END;

CREATE TRIGGER weapons_au AFTER UPDATE ON weapons BEGIN
    INSERT INTO weapons_fts(weapons_fts, rowid, name, description, effect_name)
    VALUES ('delete', old.id, old.name, old.description, old.effect_name);
    INSERT INTO weapons_fts(rowid, name, description, effect_name)
    VALUES (new.id, new.name, new.description, new.effect_name);
END;

INSERT INTO weapons_fts(rowid, name, description, effect_name)
SELECT id, name, description, effect_name FROM weapons;

CREATE INDEX weapons_type_idx ON weapons(weapon_type);
CREATE INDEX weapons_rarity_idx ON weapons(rarity);
CREATE INDEX owned_weapons_owned_idx ON owned_weapons(profile_id, owned);
