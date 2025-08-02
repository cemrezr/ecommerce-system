-- 003_create_event_logs_table.sql
CREATE TABLE IF NOT EXISTS event_logs (
                                          id SERIAL PRIMARY KEY,
                                          event_type TEXT NOT NULL,
                                          event_version TEXT NOT NULL,
                                          payload JSONB NOT NULL,
                                          status TEXT NOT NULL DEFAULT 'published', -- publishing, failed, published
                                          retry_count INT NOT NULL DEFAULT 0,
                                          created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE
    );
