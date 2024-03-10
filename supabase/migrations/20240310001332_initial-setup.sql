CREATE TABLE IF NOT EXISTS collections (
  id BIGINT PRIMARY KEY,
  name TEXT,
  weights JSONB,
  creator_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  active_at TIMESTAMP,
  retired_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS editions (
  id BIGINT PRIMARY KEY,
  name TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS collectables (
  id BIGINT PRIMARY KEY,
  name VARCHAR(100),
  collection_id BIGINT,
  creator_id BIGINT,
  rarity TEXT,
  imagepath TEXT,
  approved_at TIMESTAMP,
  approved_by BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS collectable_instances (
  id BIGINT PRIMARY KEY,
  collectable_id BIGINT,
  owner_id BIGINT,
  tags TEXT[],
  edition_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP
);

CREATE INDEX idx_collectable_instances_owner_id ON collectable_instances(owner_id);
CREATE INDEX idx_collectable_instances_created_at ON collectable_instances(created_at);

CREATE TABLE IF NOT EXISTS users (
  id BIGINT PRIMARY KEY,
  twitch_id VARCHAR(100),
  name TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_twitch_id ON users(twitch_id);

CREATE TABLE IF NOT EXISTS user_tokens (
  user_id BIGINT,
  token BYTEA PRIMARY KEY, -- Assuming `token` should be a unique identifier; if not, add a separate PRIMARY KEY column.
  access_token TEXT,
  refresh_token TEXT,
  expires_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_tokens_token ON user_tokens(token);

CREATE TABLE IF NOT EXISTS image_uploads (
  id BIGINT PRIMARY KEY,
  upload_name TEXT,
  user_id BIGINT,
  imagepath TEXT,
  uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
