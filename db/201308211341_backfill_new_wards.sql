UPDATE service_requests
SET ward_2015 = (
	SELECT ward 
	FROM ward_boundaries_2015 
	WHERE ST_ContainsProperly(boundary, ST_SetSRID(ST_Point(service_requests.long, service_requests.lat), 4326))
);