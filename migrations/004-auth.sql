CREATE TABLE IF NOT EXISTS user_auth (
  user_id INT PRIMARY KEY,
  token BINARY(32),
  access_token VARCHAR(100),
  refresh_token VARCHAR(100),
  expires_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_auth_token ON user_auth(token);
