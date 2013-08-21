package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"log"
	"time"
)

type ServiceRequest struct {
	Lat, Long                                                                                               float64
	Ward, Ward2015, Police_district                                                                                   int
	Service_request_id, Status, Service_name, Service_code, Agency_responsible, Address, Channel, Media_url string
	Requested_datetime, Updated_datetime                                                                    time.Time // FIXME: should these be proper time objects?
	Extended_attributes                                                                                     map[string]interface{}
	Notes                                                                                                   []map[string]interface{}
}

type ServiceRequestDB struct {
	InsertStmt *sql.Stmt
	UpdateStmt *sql.Stmt
	db         *sql.DB
}

func (srdb *ServiceRequestDB) Init(db *sql.DB) error {
	srdb.db = db
	srdb.SetupStmts()

	return nil
}

func (srdb *ServiceRequestDB) Close() error {
	log.Printf("Closing ServiceRequestDB database connection.")
	srdb.db.Close()

	return nil
}

func (srdb *ServiceRequestDB) SetupStmts() {
	insert, err := srdb.db.Prepare(`INSERT INTO service_requests(service_request_id,
		status, service_name, service_code, agency_responsible,
		address, requested_datetime, updated_datetime, lat, long,
		ward, police_district, media_url, channel, duplicate, parent_service_request_id, closed_datetime, notes, ward_2015)
		VALUES ($1::varchar, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19);`)

	if err != nil {
		log.Fatal("error preparing insert statement ", err)
	}
	srdb.InsertStmt = insert

	update, err := srdb.db.Prepare(`UPDATE service_requests SET
		status = $2, service_name = $3, service_code = $4, agency_responsible = $5, 
		address = $6, requested_datetime = $7, updated_datetime = $8, lat = $9, long = $10,
		ward = $11, police_district = $12, media_url = $13, channel = $14, duplicate = $15,
		parent_service_request_id = $16, updated_at = NOW(), closed_datetime = $17, notes = $18, ward_2015 = $19 
		WHERE service_request_id = $1;`)

	if err != nil {
		log.Fatal("error preparing update statement ", err)
	}
	srdb.UpdateStmt = update
}

func (req ServiceRequest) String() string {
	// pretty print SR information
	return fmt.Sprintf("%s: %s at %s %f,%f, last update %s", req.Service_request_id, req.Service_name, req.Address, req.Lat, req.Long, req.Updated_datetime)
}

func (srdb *ServiceRequestDB) Newest() (*ServiceRequest, error) {
	var newest ServiceRequest
	if err := srdb.db.QueryRow("SELECT MAX(updated_datetime) FROM service_requests;").Scan(&newest.Updated_datetime); err != nil {
		log.Print("error loading most recent SR", err)
	}
	return &newest, nil
}

func (srdb *ServiceRequestDB) Oldest() (*ServiceRequest, error) {
	var oldest ServiceRequest
	if err := srdb.db.QueryRow("SELECT MIN(updated_datetime) FROM service_requests;").Scan(&oldest.Updated_datetime); err != nil {
		log.Print("error loading oldest SR", err)
	}
	return &oldest, nil
}

func (srdb *ServiceRequestDB) Save(req *ServiceRequest) (persisted bool) { // FIXME: should return error, too
	persisted = false

	// open311 says we should always ignore a SR that does not have a SR# assigned
	if req.Service_request_id == "" {
		log.Printf("cowardly refusing to create a new SR record because of empty SR#. Request type is %s", req.Service_name)
		return persisted
	}

	// find existing record if exists
	var existing_id int
	err := srdb.db.QueryRow("SELECT id FROM service_requests WHERE service_request_id = $1", req.Service_request_id).Scan(&existing_id)
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
		stmt = srdb.InsertStmt
	} else {
		stmt = srdb.UpdateStmt
	}

	t := req.ExtractClosedDatetime()
	closed_time := pq.NullTime{Time: t, Valid: !t.IsZero()}
	notes_as_json, err := json.Marshal(req.Notes)
	if err != nil {
		log.Print("error marshaling notes to JSON: ", err)
	}
	new_ward := srdb.Ward(req, 2015)	

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
		new_ward)

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

func (req ServiceRequest) ExtractClosedDatetime() time.Time {
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

func (req ServiceRequest) PrintNotes() {
	fmt.Printf("Notes for SR %s:\n", req.Service_request_id)

	for _, note := range req.Notes {
		fmt.Printf("%+v\n", note)
	}
}

func (srdb *ServiceRequestDB) Ward(sr *ServiceRequest, year int) (ward int) {
	// given a year, return the ward containing the SR
	
	var boundaries_table string

	switch year {
	case 2013:
		boundaries_table = "ward_boundaries_2013"
	case 2015:
		boundaries_table = "ward_boundaries_2015"
	}

	query := fmt.Sprintf("SELECT ward FROM %s WHERE ST_Contains(boundary, ST_PointFromText('POINT(%f %f)', 4326))", boundaries_table, sr.Long, sr.Lat)

	err := srdb.db.QueryRow(query).Scan(&ward)
	if err != nil {
		log.Print(err)
	}

	return
}
