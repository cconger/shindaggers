ALTER TABLE knife_ownership ADD COLUMN was_subscriber BOOLEAN default false;

ALTER TABLE knife_ownership ADD COLUMN is_verified BOOLEAN default false;
