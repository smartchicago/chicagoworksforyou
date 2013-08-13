package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"log"
	"time"
)

type Open311Request struct {
	Lat, Long                                                                                               float64
	Ward, Police_district                                                                                   int
	Service_request_id, Status, Service_name, Service_code, Agency_responsible, Address, Channel, Media_url string
	Requested_datetime, Updated_datetime                                                                    string // FIXME: should these be proper time objects?
	Extended_attributes                                                                                     map[string]interface{}
	Notes                                                                                                   []map[string]interface{}
}

func (req Open311Request) String() string {
	// pretty print SR information
	return fmt.Sprintf("%s: %s at %s %f,%f, last update %s", req.Service_request_id, req.Service_name, req.Address, req.Lat, req.Long, req.Updated_datetime)
}

func (req Open311Request) Save() (persisted bool) {
	// create or update a SR

	// open311 says we should always ignore a SR that does not have a SR# assigned
	if req.Service_request_id == "" {
		log.Printf("cowardly refusing to create a new SR record because of empty SR#. Request type is %s", req.Service_name)
		return false
	}

	persisted = false

	// find existing record if exists
	var existing_id int
	err := worker.Db.QueryRow("SELECT id FROM service_requests WHERE service_request_id = $1", req.Service_request_id).Scan(&existing_id)
	switch {
	case err == sql.ErrNoRows:
		// log.Printf("did not find existing record %s", req.Service_request_id)
	case err != nil:
		log.Fatal("error searching for existing SR", err)
	default:
		persisted = true
		// log.Printf("found existing sr %s", req.Service_request_id)
	}

	var stmt *sql.Stmt

	if !persisted {
		stmt = worker.InsertStmt
	} else {
		stmt = worker.UpdateStmt
	}

	t := req.ExtractClosedDatetime()
	closed_time := pq.NullTime{Time: t, Valid: !t.IsZero()}
	notes_as_json, err := json.Marshal(req.Notes)
	if err != nil {
		log.Print("error marshaling notes to JSON: ", err)
	}

	_, err = stmt.Exec(req.Service_request_id,
		req.Status,
		req.Service_name,
		req.Service_code,
		req.Agency_responsible,
		req.Address,
		req.Requested_datetime,
		req.Updated_datetime,
		req.Lat,
		req.Long,
		req.Extended_attributes["ward"],
		req.Extended_attributes["police_district"],
		req.Media_url,
		req.Extended_attributes["channel"],
		req.Extended_attributes["duplicate"],
		req.Extended_attributes["parent_service_request_id"],
		closed_time,
		notes_as_json,
	)

	if err != nil {
		log.Printf("[error] could not update %s because %s", req.Service_request_id, err)
	} else {
		var verb string
		switch {
		case !persisted && closed_time.Time.IsZero():
			verb = "CREATED"
		case !persisted && !closed_time.Time.IsZero():
			verb = "CREATED/CLOSED"
		case persisted && closed_time.Time.IsZero():
			verb = "UPDATED"
		case persisted && !closed_time.Time.IsZero():
			verb = "UPDATED/CLOSED"
		}

		log.Printf("[%s] %s", verb, req)
		persisted = true
	}

	return persisted
}

func (req Open311Request) ExtractClosedDatetime() time.Time {
	// given an extended_attributes JSON blob, pluck out the closed time, if present
	// req.PrintNotes()

	var closed_at time.Time
	for _, note := range req.Notes {
		if note["type"] == "closed" {
			parsed_date, err := time.Parse("2006-01-02T15:04:05-07:00", note["datetime"].(string))
			if err != nil {
				log.Print("error parsing date", err)
			}
			log.Printf("SR %s closed at: %s", req, parsed_date)
			closed_at = parsed_date
		}
	}

	return closed_at
}

func (req Open311Request) PrintNotes() {
	fmt.Printf("Notes for SR %s:\n", req.Service_request_id)

	for _, note := range req.Notes {
		fmt.Printf("%+v\n", note)
	}
}
