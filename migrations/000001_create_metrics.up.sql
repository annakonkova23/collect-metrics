CREATE SCHEMA IF NOT exists storage;

CREATE TABLE IF NOT EXISTS storage.metrics_value(
    code varchar(500) not null,
    type_metric varchar(100) not null
    gauge_value double precision null
    counter_value int null
);

CREATE INDEX IF NOT EXISTS idx_metrics_value ON storage.metrics_value(code);

DO $$ BEGIN
    IF NOT EXISTS (SELECT * FROM pg_constraint 
                   WHERE conrelid = 'storage.metrics_value'::regclass AND conname = 'unq_metrics_value') THEN 
       ALTER TABLE storage.metrics_value ADD CONSTRAINT unq_metrics_value UNIQUE (code, type_metric);
    END IF;
END $$;