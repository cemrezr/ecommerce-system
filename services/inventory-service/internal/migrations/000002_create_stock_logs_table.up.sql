-- 002_create_stock_logs_table.sql
CREATE TABLE IF NOT EXISTS stock_logs (
                                          id SERIAL PRIMARY KEY,
                                          product_id BIGINT NOT NULL,
                                          change INT NOT NULL,
                                          reason TEXT NOT NULL,
                                          created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
    );
