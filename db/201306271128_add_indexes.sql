DROP INDEX IF EXISTS sr_closed_code;
DROP INDEX IF EXISTS sr_requested_code;

CREATE INDEX sr_closed_code 	ON service_requests (closed_datetime, service_code);
CREATE INDEX sr_requested_code 	ON service_requests (requested_datetime, service_code);