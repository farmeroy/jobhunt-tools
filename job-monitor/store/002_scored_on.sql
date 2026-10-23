-- +goose Up
ALTER TABLE postings ADD COLUMN scored_on DATETIME;
UPDATE postings SET scored_on = CURRENT_TIMESTAMP WHERE content_score IS NOT NULL;

-- +goose Down
ALTER TABLE postings DROP COLUMN scored_on;
