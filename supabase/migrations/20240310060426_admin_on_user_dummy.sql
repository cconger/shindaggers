ALTER TABLE user_tokens
  DROP COLUMN admin;

ALTER TABLE users
  ADD COLUMN admin BOOLEAN DEFAULT FALSE;