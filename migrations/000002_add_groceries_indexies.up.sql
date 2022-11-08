-- Filename: migrations/000001_add_groceries_indexes.up.sql
CREATE INDEX IF NOT EXISTS grocery_name_idx ON grocery USING GIN(to_tsvector('simple', name));
CREATE INDEX IF NOT EXISTS grocery_item_idx ON grocery USING GIN(to_tsvector('simple', item));