ALTER TABLE editions ADD PRIMARY KEY (id);
ALTER TABLE users ADD PRIMARY KEY (id);
ALTER TABLE users ADD PRIMARY KEY (id);

ALTER TABLE users MODIFY COLUMN id INT AUTO_INCREMENT;
ALTER TABLE editions MODIFY COLUMN id INT AUTO_INCREMENT;

CREATE TABLE IF NOT EXISTS transactions (
  id INT AUTO_INCREMENT PRIMARY KEY,
  knife_id INT,
  source INT,
  recipient INT,
  executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
);

CREATE INDEX idx_transactions_recipient ON transactions(recipient);

