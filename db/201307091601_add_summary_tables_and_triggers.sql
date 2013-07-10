DROP TABLE IF EXISTS daily_counts;

CREATE TABLE daily_counts(
	requested_date DATE NOT NULL,
	service_code VARCHAR(255) NOT NULL,
	total INTEGER NOT NULL DEFAULT 0,
	ward INTEGER NOT NULL
);

CREATE OR REPLACE FUNCTION update_daily_counts() RETURNS TRIGGER AS $update_daily_counts$
-- mostly cribbed from http://www.postgresql.org/docs/9.2/static/plpgsql-trigger.html
	DECLARE
		change	integer;
		day_to_update date;
		sr_ward integer;
		sr_service_code varchar(225);			
	BEGIN
		IF (TG_OP = 'DELETE') THEN
			-- DECREMENT
			sr_ward = NEW.ward;
			day_to_update = DATE(NEW.requested_datetime);
			sr_service_code = NEW.service_code;
			change = -1;
		ELSIF (TG_OP = 'UPDATE') THEN
			-- HANDLE CASE WHERE NON-DUP BECOMES DUP, THEN DECREMENT
			-- HANDLE CASE WHERE WARD CHANGES, DEC THEN INC
			
			sr_ward = NEW.ward;
			day_to_update = DATE(NEW.requested_datetime);
			sr_service_code = NEW.service_code;
			change = 1;
			
		ELSIF (TG_OP = 'INSERT' AND NEW.duplicate IS NULL) THEN
			-- INC IF VALID SR
			sr_ward = NEW.ward;
			day_to_update = DATE(NEW.requested_datetime);
			sr_service_code = NEW.service_code;
			change = -1;
		END IF;		

		<<insert_update>>
		LOOP
			UPDATE daily_counts
			SET total = total + change
			WHERE daily_counts.ward = sr_ward 
				AND daily_counts.requested_date = day_to_update 
				AND daily_counts.service_code = sr_service_code;
			EXIT insert_update WHEN found;

			BEGIN
				INSERT INTO daily_counts (
					requested_date,
					service_code,
					total,
					ward) 
				VALUES (
					day_to_update,
					sr_service_code,
					change,
					sr_ward);
			EXCEPTION WHEN not_null_violation THEN
				-- ignore
			END;
	
			EXIT insert_update;		
		END LOOP insert_update;
		RETURN NULL;
	END;
$update_daily_counts$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_daily_counts ON service_requests;
CREATE TRIGGER update_daily_counts
	AFTER INSERT OR UPDATE OR DELETE ON service_requests
	FOR EACH ROW EXECUTE PROCEDURE update_daily_counts();