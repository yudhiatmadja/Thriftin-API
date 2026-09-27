DROP TABLE IF EXISTS seller_applications;
DROP TABLE IF EXISTS banned_words;
ALTER TABLE users DROP COLUMN IF EXISTS seller_verified;
