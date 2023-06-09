CREATE TABLE IF NOT EXISTS knives (
  id INT AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(100),
  author_id INT,
  edition_id INT,
  rarity ENUM('Uncommon', 'Common', 'Rare', 'Ultra Rare', 'Super Rare'),
  image_name VARCHAR(100),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS editions (
  id INT PRIMARY KEY,
  name VARCHAR(100),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
  id INT PRIMARY KEY,
  twitch_id VARCHAR(100),
  twitch_name VARCHAR(50),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_twitch_id ON users(twitch_id);

CREATE TABLE IF NOT EXISTS knife_ownership (
  user_id INT,
  knife_id INT,
  instance_id INT AUTO_INCREMENT PRIMARY KEY,
  trans_type ENUM('pull', 'trade')
);

