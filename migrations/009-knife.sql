CREATE TABLE IF NOT EXISTS pullconfig  (
  community_id BIGINT,
  rarity ENUM('Uncommon', 'Common', 'Rare', 'Ultra Rare', 'Super Rare'),
  weight INT,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (community_id, rarity)
);

CREATE TABLE IF NOT EXISTS image_uploads (
  user_id INT,
  image_id BIGINT,
  path VARCHAR(100),
  uploadname VARCHAR(100),
  uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (image_id)
);
