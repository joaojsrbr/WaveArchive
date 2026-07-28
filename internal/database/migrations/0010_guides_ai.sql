CREATE TABLE character_guides (
    id TEXT PRIMARY KEY,
    character_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    source TEXT NOT NULL DEFAULT 'Aki Game',
    like_count INTEGER NOT NULL DEFAULT 0,
    language TEXT NOT NULL DEFAULT 'en',
    data_json TEXT NOT NULL DEFAULT '{}',
    synced_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(character_id) REFERENCES characters(id) ON DELETE CASCADE
);
CREATE INDEX character_guides_character_idx ON character_guides(character_id,like_count DESC);

CREATE VIRTUAL TABLE local_knowledge_fts USING fts5(
    entity_type UNINDEXED, entity_id UNINDEXED, title, content,
    tokenize='unicode61 remove_diacritics 2'
);

INSERT INTO local_knowledge_fts(entity_type,entity_id,title,content)
SELECT 'character',id,name,nickname||' '||description||' '||talent_description FROM characters;
INSERT INTO local_knowledge_fts(entity_type,entity_id,title,content)
SELECT 'weapon',id,name,description||' '||effect_name||' '||effect FROM weapons;
INSERT INTO local_knowledge_fts(entity_type,entity_id,title,content)
SELECT 'echo',id,name,skill FROM echoes;

CREATE TRIGGER local_characters_ai AFTER INSERT ON characters BEGIN
  INSERT INTO local_knowledge_fts(entity_type,entity_id,title,content)
  VALUES('character',new.id,new.name,new.nickname||' '||new.description||' '||new.talent_description);
END;
CREATE TRIGGER local_characters_au AFTER UPDATE ON characters BEGIN
  DELETE FROM local_knowledge_fts WHERE entity_type='character' AND entity_id=old.id;
  INSERT INTO local_knowledge_fts(entity_type,entity_id,title,content)
  VALUES('character',new.id,new.name,new.nickname||' '||new.description||' '||new.talent_description);
END;
CREATE TRIGGER local_weapons_ai AFTER INSERT ON weapons BEGIN
  INSERT INTO local_knowledge_fts(entity_type,entity_id,title,content)
  VALUES('weapon',new.id,new.name,new.description||' '||new.effect_name||' '||new.effect);
END;
CREATE TRIGGER local_weapons_au AFTER UPDATE ON weapons BEGIN
  DELETE FROM local_knowledge_fts WHERE entity_type='weapon' AND entity_id=old.id;
  INSERT INTO local_knowledge_fts(entity_type,entity_id,title,content)
  VALUES('weapon',new.id,new.name,new.description||' '||new.effect_name||' '||new.effect);
END;
CREATE TRIGGER local_echoes_ai AFTER INSERT ON echoes BEGIN
  INSERT INTO local_knowledge_fts(entity_type,entity_id,title,content) VALUES('echo',new.id,new.name,new.skill);
END;
CREATE TRIGGER local_echoes_au AFTER UPDATE ON echoes BEGIN
  DELETE FROM local_knowledge_fts WHERE entity_type='echo' AND entity_id=old.id;
  INSERT INTO local_knowledge_fts(entity_type,entity_id,title,content) VALUES('echo',new.id,new.name,new.skill);
END;
