DROP TRIGGER IF EXISTS trg_products_search ON products;
DROP FUNCTION IF EXISTS products_search_trigger();
DROP INDEX IF EXISTS idx_products_category;
DROP INDEX IF EXISTS idx_products_brand;
DROP INDEX IF EXISTS idx_products_price;
DROP INDEX IF EXISTS idx_products_condition;
