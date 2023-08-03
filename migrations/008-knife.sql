CREATE TABLE IF NOT EXISTS equipped (
  user_id INT,
  instance_Id INT,
  equipped_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE INDEX idx_equipped_user ON equipped(user_id);
