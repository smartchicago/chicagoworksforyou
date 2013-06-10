-- SR# must be unique

ALTER TABLE service_requests
    ADD CONSTRAINT sr_number_uniq UNIQUE(service_request_id);
