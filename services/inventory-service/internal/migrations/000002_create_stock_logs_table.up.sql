-- 002_create_stock_logs_table.sql
CREATE TABLE IF NOT EXISTS stock_logs (
                                          id SERIAL PRIMARY KEY,
                                          product_id BIGINT NOT NULL,
                                          change INT NOT NULL, -- pozitif: ekleme, negatif: çıkarma
                                          reason TEXT NOT NULL, -- örn: 'order.created', 'order.cancelled'
                                          created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
    );
