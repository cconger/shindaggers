
/*
users
  id           bigint
  name         varchar(50)
  twitch_id    varchar(50)
  twitch_login varchar(50)
  updated_at   timestamp
  created_at   timestamp
*/

CREATE TABLE IF NOT EXISTS users (
   id BIGINT PRIMARY KEY,
   name VARCHAR(100),
   twitch_id VARCHAR(100),
   twitch_login VARCHAR(50),
   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

/*
user_auth
  user_id               bigint
  token                 binary
  twitch_token          varchar
  twitch_refresh_token  varchar   
  admin                 bool
  created_at            timestamp 
  expires_at            timestamp 
*/

CREATE TABLE IF NOT EXISTS user_auth (
   user_id BIGINT PRIMARY KEY,
   token BINARY(32),
   access_token VARCHAR(100),
   refresh_token VARCHAR(100),
   expires_at TIMESTAMP,
   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_auth_token ON user_auth(token);


CREATE TABLE IF NOT EXISTS communities (
  id BIGINT PRIMARY KEY,
  name VARCHAR(100),
  owner_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

/* Users allowed to manage communities
community_admin
  user_id      bigint
  community_id bigint
*/

CREATE TABLE IF NOT EXISTS community_admin (
  user_id BIGINT,
  community_id BIGINT
);

/*
collectables
  id            bigint
  community_id  bigint
  name          varchar
  image_name    varchar   
  rarity        enum      
  author_id     bigint       
  updated_at    timestamp 
  created_at    timestamp 
  deleted       bool   
*/

CREATE TABLE IF NOT EXISTS collectables (
  id BIGINT PRIMARY KEY,
  community_id BIGINT,
  name VARCHAR(255),
  image_name VARCHAR(255),
  description TEXT,
  rarity INT,
  author_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted BOOLEAN
);

/*
lootbox
  id            bigint
  community_id  bigint
  name          varchar
  active        bool
  updated_at    timestamp 
  created_at    timestamp 
  deleted       bool   
*/

CREATE TABLE IF NOT EXISTS lootbox (
  id BIGINT PRIMARY KEY,
  community_id BIGINT,
  name VARCHAR(255),
  active BOOLEAN,
  author_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted BOOLEAN
);

/* Membership of collectables to a lootbox
lootbox_collectables
  lootbox_id     bigint
  collectable_id bigint
*/

CREATE TABLE IF NOT EXISTS lootbox_collectables (
  lootbox_id     BIGINT,
  collectable_id BIGINT
);


/* This is table for determining ownership
ownership
  id             bigint
  collectable_id bigint
  user_id        bigint
  acquired_at    timestamp
  transaction    bigint
  render_data    JSON
*/

CREATE TABLE IF NOT EXISTS ownership (
  id BIGINT PRIMARY KEY,
  collectable_id BIGINT,
  user_id BIGINT,
  aquired_at TIMESTAMP,
  transaction BIGINT,
  render_data JSON
);

/*
transactions
  id             bigint
  instance_id    bigint
  from_id        bigint
  to_id          bigint
  method         enum
*/

CREATE TABLE IF NOT EXISTS transactions (
  id BIGINT PRIMARY KEY,
  instance_id BIGINT,
  from_user_id BIGINT,
  to_user_id BIGINT,
  method enum('pull', 'gift', 'trade'),
  lore TEXT,
  execution_time TIMESTAMP
);

/* TODO add ability to have TBDC */

CREATE TABLE IF NOT EXISTS webhook (
  id BIGINT PRIMARY KEY,
  hook_token VARCHAR(64),
  user_id BIGINT,
  community_id BIGINT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  disabled BOOLEAN
);


/*
================================
# Old schema

knife_ownership
  user_id int
  knife_id int
  trans_type enum('pull','trade')
  transacted_at timestamp
  instance_id int
  was_subscriber bool
  is_verified bool
  edition_id int

users
+-------------+-----------+
| COLUMN_NAME | DATA_TYPE |
+-------------+-----------+
| admin       | tinyint   |
| created_at  | timestamp |
| id          | int       |
| lookup_name | varchar   |
| twitch_id   | varchar   |
| twitch_name | varchar   |
| updated_at  | timestamp |
+-------------+-----------+

user_auth
+---------------+-----------+
| COLUMN_NAME   | DATA_TYPE |
+---------------+-----------+
| access_token  | varchar   |
| created_at    | timestamp |
| expires_at    | timestamp |
| refresh_token | varchar   |
| token         | binary    |
| updated_at    | timestamp |
| user_id       | int       |
+---------------+-----------+

knives
+-------------+-----------+
| COLUMN_NAME | DATA_TYPE |
+-------------+-----------+
| author_id   | int       |
| created_at  | timestamp |
| deleted     | tinyint   |
| id          | int       |
| image_name  | varchar   |
| name        | varchar   |
| rarity      | enum      |
| updated_at  | timestamp |
+-------------+-----------+
*/

