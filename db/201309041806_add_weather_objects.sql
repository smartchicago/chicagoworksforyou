--
-- Name: weather_daily_stats; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

DROP TABLE IF EXISTS weather_daily_stats ;

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

DROP SEQUENCE IF EXISTS weather_storm_event_id_seq ;

CREATE SEQUENCE weather_storm_event_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: weather_storm_event; Type: TABLE; Schema: public; Owner: -; Tablespace: 
--

DROP TABLE IF EXISTS weather_storm_event ;

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
