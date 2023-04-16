ALTER TABLE knives ADD COLUMN deleted BOOLEAN DEFAULT FALSE;

CREATE INDEX idx_knife_ownership_transacted_at ON knife_ownership(trasnacted_at);
