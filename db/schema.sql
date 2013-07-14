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
-- Name: update_daily_counts(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION update_daily_counts() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
-- mostly cribbed from http://www.postgresql.org/docs/9.2/static/plpgsql-trigger.html
	DECLARE
		change	integer;
		day_to_update DATE;
        foo     integer; -- throwaway
	BEGIN
		IF (TG_OP = 'DELETE' AND OLD.duplicate IS NULL) THEN
			-- DECREMENT
            -- FIXME: this codepath is never traversed -- we never delete!
			foo = update_daily_count_bucket( DATE(OLD.requested_datetime), OLD.ward, OLD.service_code, -1 );

		ELSIF (TG_OP = 'UPDATE') THEN
		
            IF (OLD.duplicate IS NULL) THEN
        		day_to_update = DATE(OLD.requested_datetime);
        		foo = update_daily_count_bucket( day_to_update, OLD.ward, OLD.service_code, -1 );
            END IF;

    		IF (NEW.duplicate IS NULL) THEN
        		day_to_update = DATE(NEW.requested_datetime);
    			foo = update_daily_count_bucket( day_to_update, NEW.ward, NEW.service_code, 1 );
            END IF;
            
		ELSIF (TG_OP = 'INSERT' AND NEW.duplicate IS NULL) THEN
    		foo = update_daily_count_bucket(DATE(NEW.requested_datetime), NEW.ward, NEW.service_code, 1 );
		END IF;

		RETURN NULL;	
	END;
$$;


SET default_tablespace = '';

SET default_with_oids = false;

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
    closed_datetime timestamp with time zone
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
-- Name: id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY service_requests ALTER COLUMN id SET DEFAULT nextval('service_requests_id_seq'::regclass);


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
INSERT INTO schema_info VALUES ('201307091601');


--
-- PostgreSQL database dump complete
--

