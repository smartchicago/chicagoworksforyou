DROP TABLE IF EXISTS transition_areas;
CREATE TABLE transition_areas ( 
	id SERIAL, 
	boundary GEOMETRY(MULTIPOLYGON), 
	ward_2013 integer, 
	ward_2015 integer
);