-- 001_create_inventory_table.sql
CREATE TABLE IF NOT EXISTS inventory (
                                         product_id SERIAL PRIMARY KEY,
                                         product_name TEXT NOT NULL,
                                         stock INT NOT NULL DEFAULT 0,
                                         updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
    );
