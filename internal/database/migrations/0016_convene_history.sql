CREATE TABLE convene_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    player_id TEXT NOT NULL UNIQUE,
    server_id TEXT NOT NULL,
    region TEXT NOT NULL DEFAULT 'global',
    language_code TEXT NOT NULL DEFAULT 'pt',
    history_partial INTEGER NOT NULL DEFAULT 1 CHECK(history_partial IN (0,1)),
    last_imported_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE convene_pulls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    profile_id INTEGER NOT NULL,
    pool_type INTEGER NOT NULL,
    resource_id TEXT NOT NULL DEFAULT '',
    resource_type TEXT NOT NULL DEFAULT '',
    item_name TEXT NOT NULL,
    rarity INTEGER NOT NULL CHECK(rarity BETWEEN 3 AND 5),
    quantity INTEGER NOT NULL DEFAULT 1,
    obtained_at TEXT NOT NULL,
    source_index INTEGER NOT NULL DEFAULT 0,
    fingerprint TEXT NOT NULL,
    imported_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(profile_id) REFERENCES convene_profiles(id) ON DELETE CASCADE,
    UNIQUE(profile_id, fingerprint)
);

CREATE TABLE convene_pool_catalog (
    profile_id INTEGER NOT NULL,
    pool_type INTEGER NOT NULL,
    locale_key TEXT NOT NULL,
    name TEXT NOT NULL,
    short_name TEXT NOT NULL,
    kind TEXT NOT NULL,
    hard_pity INTEGER NOT NULL DEFAULT 80,
    sort_order INTEGER NOT NULL DEFAULT 0,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(profile_id, pool_type),
    FOREIGN KEY(profile_id) REFERENCES convene_profiles(id) ON DELETE CASCADE
);

CREATE INDEX convene_pulls_profile_time_idx
    ON convene_pulls(profile_id, obtained_at DESC, source_index ASC, id DESC);
CREATE INDEX convene_pulls_pool_time_idx
    ON convene_pulls(profile_id, pool_type, obtained_at DESC, source_index ASC);
CREATE INDEX convene_pulls_rarity_idx
    ON convene_pulls(profile_id, rarity, obtained_at DESC);
CREATE INDEX convene_pool_catalog_order_idx
    ON convene_pool_catalog(profile_id, sort_order, pool_type);
