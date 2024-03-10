CREATE TABLE IF NOT EXISTS user_equip_collectable_instance (
  user_id BIGINT PRIMARY KEY,
  instance_id BIGINT,
  equipped_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
