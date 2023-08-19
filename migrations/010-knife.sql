ALTER TABLE user_auth DROP PRIMARY KEY, ADD PRIMARY KEY(token);

CREATE TABLE IF NOT EXISTS fights (
  id BIGINT,
  participants JSON,
  knives JSON,
  outcomes JSON,
  event VARCHAR(100),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY(id)
);
