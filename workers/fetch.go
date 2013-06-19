package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/bmizerany/pq"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const OPEN311_API_URI = "http://311api.cityofchicago.org/open311/v2/requests.json?extensions=true&page_size=500"

type Open311Request struct {
	Lat, Long                                                                                               float64
	Ward, Police_district                                                                                   int
	Service_request_id, Status, Service_name, Service_code, Agency_responsible, Address, Channel, Media_url string
	Requested_datetime, Updated_datetime                                                                    string // FIXME: should these be proper time objects?
	Extended_attributes                                                                                     map[string]interface{}
	Notes                                                                                                   []map[string]interface{}
}

type Worker struct {
	Db        *sql.DB
	LastRunAt time.Time
}

var worker Worker

func init() {
	// open database
	db, err := sql.Open("postgres", "dbname=cwfy sslmode=disable")
	if err != nil {
		log.Fatal("Cannot open database connection", err)
	}
	worker.Db = db
}

func main() {
	defer worker.Db.Close()

	for {
		switch {
		case time.Since(worker.LastRunAt) > (30 * time.Second):
			// load requests from open311
			for _, request := range fetchRequests() {
				request.Save()
			}
			worker.LastRunAt = time.Now()
		default:
			log.Print("sleeping for 10 seconds")
			time.Sleep(10 * time.Second)
		}
	}
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
		// log.Print("error searching for existing SR", err)
	default:
		persisted = true
		// log.Printf("found existing sr %s", req.Service_request_id)
	}

	var stmt *sql.Stmt

	if !persisted {
		// create new record
		stmt, err = worker.Db.Prepare("INSERT INTO service_requests(service_request_id," +
			"status, service_name, service_code, agency_responsible, " +
			"address, requested_datetime, updated_datetime, lat, long," +
			"ward, police_district, media_url, channel, duplicate, parent_service_request_id, closed_datetime) " +
			"VALUES ($1::varchar, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17); ")
			
		        // "WHERE NOT EXISTS (SELECT 1 FROM service_requests WHERE service_request_id = $1);")
			
		if err != nil {
			log.Fatal("error preparing database insert statement", err)
		}

	} else {
		// update existing record
		stmt, err = worker.Db.Prepare("UPDATE service_requests SET " +
			"status = $2, service_name = $3, service_code = $4, agency_responsible = $5, " +
			"address = $6, requested_datetime = $7, updated_datetime = $8, lat = $9, long = $10," +
			"ward = $11, police_district = $12, media_url = $13, channel = $14, duplicate = $15, " +
			"parent_service_request_id = $16, updated_at = NOW(), closed_datetime = $17 WHERE service_request_id = $1;")

		if err != nil {
			log.Fatal("error preparing database update statement", err)
		}
	}

	tx, err := worker.Db.Begin()

	if err != nil {
		log.Fatal("error beginning transaction", err)
	}

	var closed_time pq.NullTime
	if t := req.ExtractClosedDatetime(); !t.IsZero() {
		closed_time.Time = t
	}

	_, err = tx.Stmt(stmt).Exec(req.Service_request_id,
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
	)

	if err != nil {
		log.Fatalf("could not update %s because %s", req.Service_request_id, err)
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

	err = tx.Commit()
	if err != nil {
		log.Fatal("error closing transaction", err)
	}

	return persisted

	// calculate closed time if necessary

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
			log.Printf("located closed date %s for sr: %s", parsed_date, req)
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

func fetchRequests() (requests []Open311Request) {
	// find the most recent SR that we know about in the database
	rows, err := worker.Db.Query("SELECT MAX(updated_datetime) FROM service_requests;")
	if err != nil {
		log.Fatal("error finding most recent service request", err)
	}

	last_updated_at := time.Now()
	for rows.Next() {
		if err := rows.Scan(&last_updated_at); err != nil {
			log.Print("error finding most recent SR", err)
		}

		log.Printf("most recent SR timestamp %s", last_updated_at)
	}

	// janky hack to transform the last updated timestamp into
	// a format that plays nicely with the Open311 API
	// FIXME: there HAS to be a better way to handle this.
	formatted_date_string := last_updated_at.Format(time.RFC3339)
	formatted_date_string_with_tz := formatted_date_string[0:len(formatted_date_string)-1] + "-0500" // trunc the trailing 'Z' and tack on timezone

	// construct the request URI using base params and the proper time
	open311_api_endpoint := OPEN311_API_URI + "&updated_after=" + formatted_date_string_with_tz

	log.Printf("fetching from %s", open311_api_endpoint)
	resp, err := http.Get(open311_api_endpoint)
	defer resp.Body.Close()

	if err != nil {
		log.Fatalln("error fetching from Open311 endpoint", err)
	}

	// load response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("error loading response body", err)
	}

	// parse JSON and load into an array of Open311Request objects
	err = json.Unmarshal(body, &requests)
	if err != nil {
		log.Fatal("error parsing JSON:", err)
	}

	log.Printf("received %d requests from Open311", len(requests))

	return requests
}
