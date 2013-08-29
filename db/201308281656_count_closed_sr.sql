DROP TABLE IF EXISTS daily_closed_counts;

CREATE TABLE daily_closed_counts (
    requested_date date NOT NULL,
    service_code character varying(255) NOT NULL,
    total integer DEFAULT 0 NOT NULL,
    ward integer NOT NULL
);

CREATE OR REPLACE FUNCTION update_daily_count_bucket(tbl VARCHAR, day DATE, w INTEGER, sc VARCHAR, change INTEGER) RETURNS INTEGER AS $$
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
$$ LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION update_daily_counts() RETURNS trigger
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