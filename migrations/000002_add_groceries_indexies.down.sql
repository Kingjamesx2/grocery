-- Filename: migrations/000001_add_groceries_indexes.down.sql

DROP INDEX IF EXISTS grocery_name_idx;
DROP INDEX IF EXISTS grocery_item_idx;