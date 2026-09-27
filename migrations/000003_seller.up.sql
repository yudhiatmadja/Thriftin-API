-- Seller verification + banned words filter
ALTER TABLE users ADD COLUMN IF NOT EXISTS seller_verified BOOLEAN DEFAULT false;

CREATE TABLE IF NOT EXISTS seller_applications(
  id UUID PRIMARY KEY,
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  store_name TEXT NOT NULL,
  phone TEXT NOT NULL,
  address TEXT NOT NULL,
  id_number TEXT NOT NULL,
  product_types TEXT NOT NULL,
  description TEXT,
  status TEXT NOT NULL DEFAULT 'PENDING' CHECK(status IN ('PENDING','APPROVED','REJECTED')),
  admin_note TEXT,
  created_at TIMESTAMPTZ DEFAULT now(),
  decided_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_seller_app_user_pending ON seller_applications(user_id) WHERE status='PENDING';
CREATE INDEX IF NOT EXISTS idx_seller_app_status ON seller_applications(status);

CREATE TABLE IF NOT EXISTS banned_words(
  word TEXT PRIMARY KEY,
  created_at TIMESTAMPTZ DEFAULT now()
);
INSERT INTO banned_words(word) VALUES
  ('rokok'),('vape'),('narkoba'),('sabu'),('ganja'),('ekstasi'),
  ('senjata'),('pistol'),('senapan'),('amunisi'),
  ('judi'),('togel'),('slot gacor'),('miras oplosan'),
  ('obat terlarang'),('tramadol')
ON CONFLICT DO NOTHING;
