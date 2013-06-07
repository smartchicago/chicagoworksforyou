-- adding more fields to the service_requests table

ALTER TABLE service_requests
    ADD COLUMN media_url VARCHAR(255),
    ADD COLUMN police_district INTEGER,
    ADD COLUMN ward INTEGER,
    ADD COLUMN channel VARCHAR(255),
    ADD COLUMN notes TEXT;
    