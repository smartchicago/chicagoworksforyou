package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/bmizerany/pq"
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
}

func main() {
	// open database
	db, err := sql.Open("postgres", "dbname=cwfy sslmode=disable")
	if err != nil {
		log.Fatal("Cannot open database connection", err)
	}
	defer db.Close()

	last_run_at := time.Now()

	for {
		switch {
		case time.Since(last_run_at) > (30 * time.Second):
			poll_open311(db)
			last_run_at = time.Now()
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

func fetchRequests() (requests []Open311Request) {
	db, err := sql.Open("postgres", "dbname=cwfy sslmode=disable")
	if err != nil {
		log.Fatal("Cannot open database connection", err)
	}
	defer db.Close()

	// find the most recent SR that we know about in the database
	rows, err := db.Query("SELECT MAX(updated_datetime) FROM service_requests;")
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

func poll_open311(db *sql.DB) {
	// load requests from open311
	requests := fetchRequests()

	for _, request := range requests {
		// for each request, either create or update the
		// corresponding record in the database.

		if request.Service_request_id == "" {
			log.Printf("Ignoring a request type %s because there is no SR number assigned", request.Service_name)
			continue
		}

		insert_stmt, err := db.Prepare("INSERT INTO service_requests(service_request_id," +
			"status, service_name, service_code, agency_responsible, " +
			"address, requested_datetime, updated_datetime, lat, long," +
			"ward, police_district, media_url, channel, duplicate, parent_service_request_id) " +
			"SELECT $1::varchar, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16 " +
			"WHERE NOT EXISTS (SELECT 1 FROM service_requests WHERE service_request_id = $1);")

		if err != nil {
			log.Fatal("error preparing database insert statement", err)
		}

		update_stmt, err := db.Prepare("UPDATE service_requests SET " +
			"status = $2, service_name = $3, service_code = $4, agency_responsible = $5, " +
			"address = $6, requested_datetime = $7, updated_datetime = $8, lat = $9, long = $10," +
			"ward = $11, police_district = $12, media_url = $13, channel = $14, duplicate = $15, " +
			"parent_service_request_id = $16, updated_at = NOW() WHERE service_request_id = $1;")

		if err != nil {
			log.Fatal("error preparing database update statement", err)
		}

		tx, err := db.Begin()

		if err != nil {
			log.Fatal("error beginning transaction", err)
		}

		_, err = tx.Stmt(update_stmt).Exec(request.Service_request_id,
			request.Status,
			request.Service_name,
			request.Service_code,
			request.Agency_responsible,
			request.Address,
			request.Requested_datetime,
			request.Updated_datetime,
			request.Lat,
			request.Long,
			request.Extended_attributes["ward"],
			request.Extended_attributes["police_district"],
			request.Media_url,
			request.Extended_attributes["channel"],
			request.Extended_attributes["duplicate"],
			request.Extended_attributes["parent_service_request_id"])

		if err != nil {
			log.Fatalf("could not update %s because %s", request.Service_request_id, err)
		}

		_, err = tx.Stmt(insert_stmt).Exec(request.Service_request_id,
			request.Status,
			request.Service_name,
			request.Service_code,
			request.Agency_responsible,
			request.Address,
			request.Requested_datetime,
			request.Updated_datetime,
			request.Lat,
			request.Long,
			request.Extended_attributes["ward"],
			request.Extended_attributes["police_district"],
			request.Media_url,
			request.Extended_attributes["channel"],
			request.Extended_attributes["duplicate"],
			request.Extended_attributes["parent_service_request_id"])

		if err != nil {
			log.Fatalf("could not save %s because %s", request.Service_request_id, err)
		} else {
			log.Printf("saved SR %s", request)
		}

		err = tx.Commit()
		if err != nil {
			log.Fatal("error closing transaction", err)
		}
	}
}
