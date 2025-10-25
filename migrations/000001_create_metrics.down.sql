DROP INDEX IF EXISTS idx_metrics_value;
ALTER TABLE storage.metrics_value DROP CONSTRAINT IF EXISTS unq_metrics_value;
DROP TABLE IF EXISTS storage.metrics_value;