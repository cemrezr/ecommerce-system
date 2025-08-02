DROP INDEX IF EXISTS unique_inventory_event;

ALTER TABLE stock_logs
DROP COLUMN IF EXISTS order_id;
