DROP TABLE IF EXISTS daily_closed_counts;

CREATE TABLE daily_closed_counts (
    requested_date date NOT NULL,
    service_code character varying(255) NOT NULL,
    total integer DEFAULT 0 NOT NULL,
    ward integer NOT NULL
);


CREATE OR REPLACE FUNCTION update_daily_count_bucket (tbl VARCHAR, day DATE, w INTEGER, sc VARCHAR, change INTEGER) RETURNS INTEGER LANGUAGE plpgsql
    AS $$
    BEGIN
        <<insert_update>>
		LOOP
            UPDATE tbl
    		SET total = total + change
    		WHERE tbl.ward = w 
    			AND tbl.requested_date = day 
    			AND tbl.service_code = sc;
            
            EXIT insert_update WHEN found;
            	
			BEGIN
				INSERT INTO tbl ( requested_date, service_code, total, ward) 
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