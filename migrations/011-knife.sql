CREATE TABLE IF NOT EXISTS fight_outcomes (
  fight_id BIGINT,
  user_id BIGINT,
  collectable_id BIGINT,
  outcome ENUM('win', 'loss', 'draw'),
  PRIMARY KEY(fight_id, user_id)
);

CREATE INDEX fight_outcomes_user_id ON fight_outcomes(user_id);
CREATE INDEX fight_outcomes_collectable_id ON fight_outcomes(collectable_id);
