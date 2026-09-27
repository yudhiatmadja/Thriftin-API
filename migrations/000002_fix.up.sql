-- Fix search_vector: trigger + backfill + indexes
CREATE OR REPLACE FUNCTION products_search_trigger() RETURNS trigger AS $$
BEGIN
  NEW.search_vector := to_tsvector('simple', coalesce(NEW.title,'') || ' ' || coalesce(NEW.description,'') || ' ' || coalesce(NEW.color,'') || ' ' || coalesce(NEW.material,''));
  RETURN NEW;
END $$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_products_search ON products;
CREATE TRIGGER trg_products_search BEFORE INSERT OR UPDATE OF title, description, color, material ON products
FOR EACH ROW EXECUTE FUNCTION products_search_trigger();

UPDATE products SET search_vector = to_tsvector('simple', coalesce(title,'') || ' ' || coalesce(description,'') || ' ' || coalesce(color,'') || ' ' || coalesce(material,''))
WHERE search_vector IS NULL;

CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_brand ON products(brand_id);
CREATE INDEX IF NOT EXISTS idx_products_price ON products(price);
CREATE INDEX IF NOT EXISTS idx_products_condition ON products(condition);
