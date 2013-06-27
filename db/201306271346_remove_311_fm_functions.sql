DROP TABLE IF EXISTS update_log;

DROP FUNCTION IF EXISTS ward_summary (ward_number text, start_date timestamp without time zone, end_date timestamp without time zone);
DROP FUNCTION IF EXISTS ward_summary_minimal (ward_number text, start_date timestamp without time zone, end_date timestamp without time zone);
DROP FUNCTION IF EXISTS ward_summary (start_date timestamp without time zone, end_date timestamp without time zone);
