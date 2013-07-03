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

SET default_tablespace = '';

SET default_with_oids = false;

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
-- PostgreSQL database dump complete
--

