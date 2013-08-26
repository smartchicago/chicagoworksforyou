CREATE EXTENSION IF NOT EXISTS postgis;

DROP TABLE IF EXISTS ward_boundaries_2013;
CREATE TABLE ward_boundaries_2013 (
	ward integer,	
	boundary GEOMETRY(MULTIPOLYGON)
);

DROP TABLE IF EXISTS ward_boundaries_2015;
CREATE TABLE ward_boundaries_2015 (
	ward integer,	
	boundary GEOMETRY(MULTIPOLYGON)
);

