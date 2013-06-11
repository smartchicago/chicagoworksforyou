-- adding more fields to the service_requests table

ALTER TABLE service_requests
    ADD COLUMN duplicate VARCHAR(40),
    ADD COLUMN parent_service_request_id VARCHAR(40);
    