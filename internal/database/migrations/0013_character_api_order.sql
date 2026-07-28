ALTER TABLE characters ADD COLUMN api_order INTEGER NOT NULL DEFAULT 0;
CREATE INDEX characters_api_order_idx ON characters(api_order);
