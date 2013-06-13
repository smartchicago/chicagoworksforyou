--
-- PostgreSQL database dump
--

SET statement_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

--
-- Name: cwfy; Type: DATABASE; Schema: -; Owner: -
--

CREATE DATABASE cwfy WITH TEMPLATE = template0 ENCODING = 'UTF8' LC_COLLATE = 'en_US.UTF-8' LC_CTYPE = 'en_US.UTF-8';


\connect cwfy

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
-- Name: ward_summary(timestamp without time zone, timestamp without time zone); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION ward_summary(start_date timestamp without time zone, end_date timestamp without time zone) RETURNS SETOF record
    LANGUAGE plpgsql
    AS $$

  DECLARE
  rec RECORD;

  BEGIN

  FOR i IN 1..50 LOOP -- OUTER

  FOR rec in SELECT ward,
                  opened_requests, 
                  closed_requests, 
                  tardy_requests
           FROM ward_summary_minimal(trim(to_char(i, '99')), start_date, end_date)
           AS (ward text,
               opened_requests int,
               closed_requests int,
               tardy_requests int)
  LOOP -- INNER
    RETURN NEXT rec;
  END LOOP; -- INNER

  END LOOP; -- OUTER
  
  RETURN;

  END;
$$;


--
-- Name: ward_summary(text, timestamp without time zone, timestamp without time zone); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION ward_summary(ward_number text, start_date timestamp without time zone, end_date timestamp without time zone) RETURNS record
    LANGUAGE plpgsql
    AS $$

  DECLARE

  opened_requests integer;
  closed_requests integer;
  tardy_requests integer;
  days_to_close_requests_avg double precision;
  request_time_bins text;
  request_time_bin_morning integer;
  request_time_bin_afternoon integer;
  request_time_bin_night integer;
  request_time_bin_sunday integer; -- 0
  request_time_bin_monday integer; -- 1
  request_time_bin_tuesday integer; -- 2
  request_time_bin_wednesday integer; -- 3
  request_time_bin_thursday integer; -- 4
  request_time_bin_friday integer; -- 5
  request_time_bin_saturday integer; -- 6
  request_types text;
  request_types_ret_row RECORD;
  ret RECORD;

  BEGIN

  -- COUNT REQUESTS OPENED DURING RANGE
  SELECT count(*)
    INTO opened_requests
    FROM service_requests 
    WHERE ward = ward_number
      AND requested_datetime >= start_date
      AND requested_datetime < end_date;

  -- COUNT REQUESTS CLOSED DURING RANGE
  SELECT count(*) 
    INTO closed_requests 
    FROM service_requests
    WHERE ward = ward_number
      AND closed_datetime >= start_date 
      AND closed_datetime < end_date;

  -- COUNT REQUESTS OPEN > 1 MONTH (28 DAYS)
  SELECT count(*) from service_requests
    INTO tardy_requests
    WHERE status = 'open' 
      AND requested_datetime >= start_date 
      AND requested_datetime < end_date
      AND ward = ward_number
      AND extract(DAY from now() - requested_datetime) > 28;

  -- CALCULATE AVG DAYS CLOSED REQUESTS TAKE TO REACH CLOSED STATE
  SELECT avg(extract(DAY from closed_datetime - requested_datetime))
  INTO days_to_close_requests_avg
  FROM service_requests
  WHERE status = 'closed'
    AND ward = ward_number
    AND requested_datetime >= start_date 
    AND requested_datetime < end_date;

  -- BIN REQUEST TIMES INTO TIME OF DAY BUCKETS

  -- GET MORNING BIN
  SELECT count(*)
  INTO request_time_bin_morning
  FROM service_requests 
  WHERE extract(hour from requested_datetime) >= 0
    AND extract(hour from requested_datetime) < 11
    AND ward = ward_number
    AND requested_datetime >= start_date 
    AND requested_datetime < end_date;

  -- GET AFTERNOON BIN
  SELECT count(*)
  INTO request_time_bin_afternoon
  FROM service_requests 
  WHERE extract(hour from requested_datetime) >= 11
    AND extract(hour from requested_datetime) < 17
    AND ward = ward_number
    AND requested_datetime >= start_date 
    AND requested_datetime < end_date;    

  -- GET NIGHT BIN
  SELECT count(*)
  INTO request_time_bin_night
  FROM service_requests 
  WHERE extract(hour from requested_datetime) >= 17
    AND extract(hour from requested_datetime) < 24
    AND ward = ward_number
    AND requested_datetime >= start_date 
    AND requested_datetime < end_date;

  -- GET SUNDAY BIN
  SELECT count(*)
  INTO request_time_bin_sunday
  FROM service_requests
  WHERE extract(DOW from requested_datetime) = 0
    AND ward = ward_number
    AND requested_datetime >= start_date 
    AND requested_datetime < end_date;

  -- GET MONDAY BIN
  SELECT count(*)
  INTO request_time_bin_monday
  FROM service_requests
  WHERE extract(DOW from requested_datetime) = 1
    AND ward = ward_number
    AND requested_datetime >= start_date 
    AND requested_datetime < end_date;

  -- GET TUESDAY BIN
  SELECT count(*)
  INTO request_time_bin_tuesday
  FROM service_requests
  WHERE extract(DOW from requested_datetime) = 2
    AND ward = ward_number
    AND requested_datetime >= start_date 
    AND requested_datetime < end_date;

  -- GET WEDNESDAY BIN
  SELECT count(*)
  INTO request_time_bin_wednesday
  FROM service_requests
  WHERE extract(DOW from requested_datetime) = 3
    AND ward = ward_number
    AND requested_datetime >= start_date 
    AND requested_datetime < end_date;

  -- GET THURSDAY BIN
  SELECT count(*)
  INTO request_time_bin_thursday
  FROM service_requests
  WHERE extract(DOW from requested_datetime) = 4
    AND ward = ward_number
    AND requested_datetime >= start_date 
    AND requested_datetime < end_date;

  -- GET FRIDAY BIN
  SELECT count(*)
  INTO request_time_bin_friday
  FROM service_requests
  WHERE extract(DOW from requested_datetime) = 5
    AND ward = ward_number
    AND requested_datetime >= start_date 
    AND requested_datetime < end_date;

  -- GET SATURDAY BIN
  SELECT count(*)
  INTO request_time_bin_saturday
  FROM service_requests
  WHERE extract(DOW from requested_datetime) = 6
    AND ward = ward_number
    AND requested_datetime >= start_date 
    AND requested_datetime < end_date;        

 -- PACKAGE TIME BINS
 request_time_bins := '{"morning":' || request_time_bin_morning || 
                      ',"afternoon":' || request_time_bin_afternoon || 
                      ',"night":' || request_time_bin_night || 
                      ',"days": {' ||
                      '    "sunday":' || request_time_bin_sunday || ',' ||
                      '    "monday":' || request_time_bin_monday || ',' ||
                      '    "tuesday":' || request_time_bin_tuesday || ',' ||
                      '    "wednesday":' || request_time_bin_wednesday || ',' ||
                      '    "thursday":' || request_time_bin_thursday || ',' ||
                      '    "friday":' || request_time_bin_friday || ',' ||
                      '    "saturday":' || request_time_bin_saturday ||
                      '  }' ||
                      '}';

  -- GET REQUEST COUNTS
  request_types := '[';
  FOR request_types_ret_row IN SELECT count(*) as count, service_name 
  FROM service_requests
  WHERE ward = ward_number
    AND requested_datetime >= start_date
    AND requested_datetime < end_date
  GROUP BY service_name
  ORDER BY count(*)  DESC
  LOOP
    request_types := request_types || '{"type":"' || request_types_ret_row.service_name || '", "count":' || request_types_ret_row.count || '},';
  END LOOP;
  request_types := trim(trailing ',' from request_types) || ']';
  
  -- PACKAGE FOR SHIPPING
  SELECT opened_requests, 
         closed_requests, 
         tardy_requests,
         days_to_close_requests_avg,
         request_time_bins,
         request_types
  INTO ret;

  -- SEND IT
  RETURN ret;

  END;
$$;


--
-- Name: ward_summary_minimal(text, timestamp without time zone, timestamp without time zone); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION ward_summary_minimal(ward_number text, start_date timestamp without time zone, end_date timestamp without time zone) RETURNS record
    LANGUAGE plpgsql
    AS $$

  DECLARE

  opened_requests integer;
  closed_requests integer;
  tardy_requests integer;
  ret RECORD;

  BEGIN

  -- COUNT REQUESTS OPENED DURING RANGE
  SELECT count(*)
    INTO opened_requests
    FROM service_requests 
    WHERE ward = ward_number
      AND requested_datetime >= start_date
      AND requested_datetime < end_date;

  -- COUNT REQUESTS CLOSED DURING RANGE
  SELECT count(*) 
    INTO closed_requests 
    FROM service_requests
    WHERE ward = ward_number
      AND closed_datetime >= start_date 
      AND closed_datetime < end_date;

  -- COUNT REQUESTS OPEN > 1 MONTH (28 DAYS)
  SELECT count(*) from service_requests
    INTO tardy_requests
    WHERE status = 'open' 
      AND requested_datetime >= start_date 
      AND requested_datetime < end_date
      AND ward = ward_number
      AND extract(DAY from now() - requested_datetime) > 28;
  
  -- PACKAGE FOR SHIPPING
  SELECT ward_number as ward,
         opened_requests, 
         closed_requests, 
         tardy_requests
  INTO ret;

  -- SEND IT
  RETURN ret;

  END;
$$;


SET default_tablespace = '';

SET default_with_oids = false;

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
    requested_datetime timestamp without time zone,
    updated_datetime timestamp without time zone,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    lat double precision,
    long double precision,
    media_url character varying(255),
    police_district integer,
    ward integer,
    channel character varying(255),
    notes text,
    duplicate character varying(40),
    parent_service_request_id character varying(40)
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
-- Name: update_log; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

CREATE TABLE update_log (
    last_run_at timestamp with time zone NOT NULL,
    notes character varying(100)
);


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
-- Name: update_log_pkey; Type: CONSTRAINT; Schema: public; Owner: -; Tablespace: 
--

ALTER TABLE ONLY update_log
    ADD CONSTRAINT update_log_pkey PRIMARY KEY (last_run_at);


--
-- PostgreSQL database dump complete
--

