-- +goose Up
-- This is the baseline migration. The schema was created by XORM Sync2.
-- No-op: existing tables are already in place.
SELECT 1;

-- +goose Down
-- Baseline cannot be rolled back.
SELECT 1;
