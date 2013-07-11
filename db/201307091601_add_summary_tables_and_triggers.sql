DROP TABLE IF EXISTS daily_counts;

CREATE TABLE daily_counts(
	requested_date DATE NOT NULL,
	service_code VARCHAR(255) NOT NULL,
	total INTEGER NOT NULL DEFAULT 0,
	ward INTEGER NOT NULL
);

-- index for quick lookups of dates 
DROP INDEX IF EXISTS dc_request_date,dc_ward,dc_service_code;
CREATE INDEX dc_request_date ON daily_counts(requested_date);
CREATE INDEX dc_ward ON daily_counts(ward);
CREATE INDEX dc_service_code ON daily_counts(service_code);

CREATE OR REPLACE FUNCTION update_daily_counts() RETURNS TRIGGER AS $update_daily_counts$
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
$update_daily_counts$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_daily_count_bucket (day DATE, w INTEGER, sc VARCHAR, change INTEGER) RETURNS INTEGER AS $$
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
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_daily_counts ON service_requests;
CREATE TRIGGER update_daily_counts
	AFTER INSERT OR UPDATE OR DELETE ON service_requests
	FOR EACH ROW EXECUTE PROCEDURE update_daily_counts();