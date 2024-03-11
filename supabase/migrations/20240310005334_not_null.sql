ALTER TABLE collections
  ALTER COLUMN name SET NOT NULL,
  ALTER COLUMN created_at SET NOT NULL,
  ALTER COLUMN updated_at SET NOT NULL,
  ALTER COLUMN creator_id SET NOT NULL;

ALTER TABLE editions
  ALTER COLUMN created_at SET NOT NULL,
  ALTER COLUMN name SET NOT NULL;

ALTER TABLE collectables
  ALTER COLUMN name SET NOT NULL,
  ALTER COLUMN creator_id SET NOT NULL,
  ALTER COLUMN rarity SET NOT NULL,
  ALTER COLUMN created_at SET NOT NULL,
  ALTER COLUMN imagepath SET NOT NULL;

ALTER TABLE collectable_instances
  ALTER COLUMN collectable_id SET NOT NULL,
  ALTER COLUMN owner_id SET NOT NULL,
  ALTER COLUMN created_at SET NOT NULL,
  ALTER COLUMN edition_id SET NOT NULL;

ALTER TABLE users
  ALTER COLUMN created_at SET NOT NULL,
  ALTER COLUMN name SET NOT NULL;

ALTER TABLE user_tokens
  ALTER COLUMN created_at SET NOT NULL,
  ALTER COLUMN updated_at SET NOT NULL,
  ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE image_uploads
  ALTER COLUMN user_id SET NOT NULL,
  ALTER COLUMN imagepath SET NOT NULL,
  ALTER COLUMN uploaded_at SET NOT NULL;