DROP TABLE IF EXISTS service_requests;
CREATE TABLE service_requests (
	id SERIAL,
	service_request_id varchar(12),
	status varchar(12),
	service_name varchar(255),
	service_code varchar(255), 
	agency_responsible varchar(255),
	address varchar(255),
	requested_datetime timestamp,
	updated_datetime timestamp,
	created_at timestamp DEFAULT current_timestamp,
	updated_at timestamp DEFAULT current_timestamp,
	lat double precision,
	long double precision,
	PRIMARY KEY (id)		
);