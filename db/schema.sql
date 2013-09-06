--
-- PostgreSQL database dump
--

SET statement_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


--
-- Name: postgis; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS postgis WITH SCHEMA public;


--
-- Name: EXTENSION postgis; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION postgis IS 'PostGIS geometry, geography, and raster spatial types and functions';


SET search_path = public, pg_catalog;

--
-- Name: update_daily_count_bucket(date, integer, character varying, integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION update_daily_count_bucket(day date, w integer, sc character varying, change integer) RETURNS integer
    LANGUAGE plpgsql
    AS $$
    BEGIN
        <<insert_update>>
		LOOP
            UPDATE daily_counts
    		SET total = total + change
    		WHERE daily_counts.ward = w 
    			AND daily_counts.requested_date = day 
    			AND daily_counts.service_code = sc;
            
            EXIT insert_update WHEN found;
            	
			BEGIN
				INSERT INTO daily_counts ( requested_date, service_code, total, ward) 
				VALUES ( day, sc, change, w);
				
			EXCEPTION WHEN not_null_violation THEN
				-- ignore
			END;

			EXIT insert_update;		
		END LOOP insert_update;

        RETURN NULL;
    END;
$$;


--
-- Name: update_daily_count_bucket(character varying, date, integer, character varying, integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION update_daily_count_bucket(tbl character varying, day date, w integer, sc character varying, change integer) RETURNS integer
    LANGUAGE plpgsql
    AS $_$
    DECLARE
        updated integer;
    BEGIN
        <<insert_update>>
        LOOP
            EXECUTE 'UPDATE ' || tbl || ' SET total = total + $1 WHERE ward = $2 AND requested_date = $3 AND service_code = $4;' USING change, w, day, sc;            
            GET DIAGNOSTICS updated = ROW_COUNT;
            EXIT insert_update WHEN updated != 0;
             
            BEGIN
                IF (change > 0) THEN
                    EXECUTE 'INSERT INTO ' || tbl || ' (requested_date, service_code, total, ward) VALUES ($1, $2, $3, $4);' USING day, sc, change, w;
                END IF;
                
                EXCEPTION WHEN not_null_violation THEN
                    -- ignore
            END;
            EXIT insert_update;     
        END LOOP insert_update;
    RETURN NULL;
    END;
$_$;


--
-- Name: update_daily_counts(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION update_daily_counts() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
    -- mostly cribbed from http://www.postgresql.org/docs/9.2/static/plpgsql-trigger.html
    DECLARE
        change  integer;
        day_to_update DATE;
        foo     integer; -- throwaway
    BEGIN
        IF (TG_OP = 'DELETE' AND OLD.duplicate IS NULL) THEN
            -- DECREMENT
            -- FIXME: this codepath is never traversed -- we never delete!
            foo = update_daily_count_bucket('daily_counts', DATE(OLD.requested_datetime), OLD.ward, OLD.service_code, -1 );

            IF (OLD.closed_datetime IS NOT NULL) THEN
                foo = update_daily_count_bucket('daily_closed_counts', DATE(OLD.closed_datetime), OLD.ward, OLD.service_code, -1 );
            END IF;

     ELSIF (TG_OP = 'UPDATE') THEN
     
            -- if the OLD record was not a duplicate, decrement the day is was marked as opened
            -- if the NEW record is not a duplicate, increment the day is is marked as opened
            -- this means that if a record changes state to or from duplicate, we maintain
            -- sane counts for that request. In most cases, it'll come to us as a duplicate
            -- and stay a duplicate, rendering this all moot.
            
            IF (OLD.duplicate IS NULL) THEN
                day_to_update = DATE(OLD.requested_datetime);
                foo = update_daily_count_bucket( 'daily_counts', day_to_update, OLD.ward, OLD.service_code, -1 );
             
                IF (OLD.closed_datetime IS NOT NULL) THEN
                    foo = update_daily_count_bucket('daily_closed_counts', DATE(OLD.closed_datetime), OLD.ward, OLD.service_code, -1 );
                END IF;
            END IF;

         IF (NEW.duplicate IS NULL) THEN
             day_to_update = DATE(NEW.requested_datetime);
             foo = update_daily_count_bucket( 'daily_counts', day_to_update, NEW.ward, NEW.service_code, 1 );
             
                IF (NEW.closed_datetime IS NOT NULL) THEN
                    foo = update_daily_count_bucket('daily_closed_counts', DATE(NEW.closed_datetime), NEW.ward, NEW.service_code, 1 );
                END IF;
             
            END IF;
            
     ELSIF (TG_OP = 'INSERT' AND NEW.duplicate IS NULL) THEN
            foo = update_daily_count_bucket('daily_counts', DATE(NEW.requested_datetime), NEW.ward, NEW.service_code, 1 );
         
            IF (NEW.closed_datetime IS NOT NULL) THEN
                foo = update_daily_count_bucket('daily_closed_counts', DATE(NEW.closed_datetime), NEW.ward, NEW.service_code, 1 );
            END IF;
     END IF;

     RETURN NULL;    
 END;
$$;


SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: daily_closed_counts; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE daily_closed_counts (
    requested_date date NOT NULL,
    service_code character varying(255) NOT NULL,
    total integer DEFAULT 0 NOT NULL,
    ward integer NOT NULL
);


--
-- Name: daily_counts; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE daily_counts (
    requested_date date NOT NULL,
    service_code character varying(255) NOT NULL,
    total integer DEFAULT 0 NOT NULL,
    ward integer NOT NULL
);


--
-- Name: schema_info; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE schema_info (
    version character varying(12)
);


--
-- Name: service_requests; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE service_requests (
    id integer NOT NULL,
    service_request_id character varying(12),
    status character varying(12),
    service_name character varying(255),
    service_code character varying(255),
    agency_responsible character varying(255),
    address character varying(255),
    requested_datetime timestamp with time zone,
    updated_datetime timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    lat double precision,
    long double precision,
    media_url character varying(255),
    police_district integer,
    ward integer,
    channel character varying(255),
    notes text,
    duplicate character varying(40),
    parent_service_request_id character varying(40),
    closed_datetime timestamp with time zone,
    ward_2015 integer,
    transition_area_id integer
);


--
-- Name: service_requests_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE service_requests_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: service_requests_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE service_requests_id_seq OWNED BY service_requests.id;


--
-- Name: transition_areas; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE transition_areas (
    id integer NOT NULL,
    boundary geometry(MultiPolygon),
    ward_2013 integer,
    ward_2015 integer
);


--
-- Name: transition_areas_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE transition_areas_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: transition_areas_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE transition_areas_id_seq OWNED BY transition_areas.id;


--
-- Name: ward_boundaries_2013; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE ward_boundaries_2013 (
    ward integer,
    boundary geometry(MultiPolygon)
);


--
-- Name: ward_boundaries_2015; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE ward_boundaries_2015 (
    ward integer,
    boundary geometry(MultiPolygon)
);


--
-- Name: weather_daily_stats; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE weather_daily_stats (
    weather_date date NOT NULL,
    high_temp_f double precision,
    low_temp_f double precision,
    precip_in double precision
);


--
-- Name: COLUMN weather_daily_stats.weather_date; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN weather_daily_stats.weather_date IS 'date for this weather data';


--
-- Name: COLUMN weather_daily_stats.high_temp_f; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN weather_daily_stats.high_temp_f IS 'day''s high temperature, in farenheit';


--
-- Name: COLUMN weather_daily_stats.low_temp_f; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN weather_daily_stats.low_temp_f IS 'day''s low temperature, in farenheit';


--
-- Name: COLUMN weather_daily_stats.precip_in; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN weather_daily_stats.precip_in IS 'day''s precipitation, in inches';


--
-- Name: weather_storm_event_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE weather_storm_event_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: weather_storm_event; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE weather_storm_event (
    id integer DEFAULT nextval('weather_storm_event_id_seq'::regclass) NOT NULL,
    event_date date NOT NULL,
    event_id integer NOT NULL,
    event_type character varying(255) NOT NULL
);


--
-- Name: TABLE weather_storm_event; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE weather_storm_event IS 'select data drawn from the NOAA NCDC Storm Events database';


--
-- Name: COLUMN weather_storm_event.id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN weather_storm_event.id IS 'internal, database-only ID for this event';


--
-- Name: COLUMN weather_storm_event.event_date; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN weather_storm_event.event_date IS 'start date of the event; based on the source data''s "begin_date" field';


--
-- Name: COLUMN weather_storm_event.event_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN weather_storm_event.event_id IS 'event ID; based on the source data''s "event_id" field';


--
-- Name: COLUMN weather_storm_event.event_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN weather_storm_event.event_type IS 'type of event; based on the source data''s "event_type" field';


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY service_requests ALTER COLUMN id SET DEFAULT nextval('service_requests_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY transition_areas ALTER COLUMN id SET DEFAULT nextval('transition_areas_id_seq'::regclass);


--
-- Name: service_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: -; Tablespace: 
--

ALTER TABLE ONLY service_requests
    ADD CONSTRAINT service_requests_pkey PRIMARY KEY (id);


--
-- Name: sr_number_uniq; Type: CONSTRAINT; Schema: public; Owner: -; Tablespace: 
--

ALTER TABLE ONLY service_requests
    ADD CONSTRAINT sr_number_uniq UNIQUE (service_request_id);


--
-- Name: dc_request_date; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX dc_request_date ON daily_counts USING btree (requested_date);


--
-- Name: dc_service_code; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX dc_service_code ON daily_counts USING btree (service_code);


--
-- Name: dc_ward; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX dc_ward ON daily_counts USING btree (ward);


--
-- Name: sr_closed_code; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX sr_closed_code ON service_requests USING btree (closed_datetime, service_code);


--
-- Name: sr_requested_code; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX sr_requested_code ON service_requests USING btree (requested_datetime, service_code);


--
-- Name: sr_requested_datetime; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX sr_requested_datetime ON service_requests USING btree (requested_datetime);


--
-- Name: sr_updated_datetime; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX sr_updated_datetime ON service_requests USING btree (updated_datetime);


--
-- Name: transition_areas_boundary_gist; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX transition_areas_boundary_gist ON transition_areas USING gist (boundary);


--
-- Name: ward_boundaries_2013_boundary_gist; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX ward_boundaries_2013_boundary_gist ON ward_boundaries_2013 USING gist (boundary);


--
-- Name: ward_boundaries_2015_boundary_gist; Type: INDEX; Schema: public; Owner: -; Tablespace: 
--

CREATE INDEX ward_boundaries_2015_boundary_gist ON ward_boundaries_2015 USING gist (boundary);


--
-- Name: geometry_columns_delete; Type: RULE; Schema: public; Owner: -
--

CREATE RULE geometry_columns_delete AS ON DELETE TO geometry_columns DO INSTEAD NOTHING;


--
-- Name: geometry_columns_insert; Type: RULE; Schema: public; Owner: -
--

CREATE RULE geometry_columns_insert AS ON INSERT TO geometry_columns DO INSTEAD NOTHING;


--
-- Name: geometry_columns_update; Type: RULE; Schema: public; Owner: -
--

CREATE RULE geometry_columns_update AS ON UPDATE TO geometry_columns DO INSTEAD NOTHING;


--
-- Name: update_daily_counts; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_daily_counts AFTER INSERT OR DELETE OR UPDATE ON service_requests FOR EACH ROW EXECUTE PROCEDURE update_daily_counts();


--
-- PostgreSQL database dump complete
--

--
-- PostgreSQL database dump
--

SET statement_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

SET search_path = public, pg_catalog;

--
-- Data for Name: schema_info; Type: TABLE DATA; Schema: public; Owner: cgansen
--

INSERT INTO schema_info VALUES ('201306061151');
INSERT INTO schema_info VALUES ('201306071651');
INSERT INTO schema_info VALUES ('201306071725');
INSERT INTO schema_info VALUES ('201306091221');
INSERT INTO schema_info VALUES ('201306161511');
INSERT INTO schema_info VALUES ('201306241712');
INSERT INTO schema_info VALUES ('201306251155');
INSERT INTO schema_info VALUES ('201306271128');
INSERT INTO schema_info VALUES ('201306271346');
INSERT INTO schema_info VALUES ('201307081428');
INSERT INTO schema_info VALUES ('201308211341');
INSERT INTO schema_info VALUES ('201307091601');
INSERT INTO schema_info VALUES ('201308271926');
INSERT INTO schema_info VALUES ('201308201558');
INSERT INTO schema_info VALUES ('201308211328');
INSERT INTO schema_info VALUES ('201308211659');
INSERT INTO schema_info VALUES ('201308281656');
INSERT INTO schema_info VALUES ('201309041806');


--
-- PostgreSQL database dump complete
--

