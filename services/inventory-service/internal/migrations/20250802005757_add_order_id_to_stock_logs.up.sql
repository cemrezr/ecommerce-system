ALTER TABLE stock_logs
    ADD COLUMN order_id BIGINT;

CREATE UNIQUE INDEX unique_inventory_event
    ON stock_logs (product_id, reason, order_id)
    WHERE order_id IS NOT NULL;
