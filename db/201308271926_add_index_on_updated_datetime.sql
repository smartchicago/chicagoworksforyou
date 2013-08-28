DROP INDEX IF EXISTS sr_updated_datetime;

CREATE INDEX sr_updated_datetime 
ON service_requests 
USING btree (updated_datetime);