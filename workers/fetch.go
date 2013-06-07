package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/bmizerany/pq"
	"io/ioutil"
	"log"
	"net/http"
)

const OPEN311_API_URI = "http://311api.cityofchicago.org/open311/v2/requests.json?extensions=true&page_size=100"

type Open311Request struct {
	Lat, Long                                                                                                      float64
	Ward, Police_district                                                                                          int
	Service_request_id, Status, Service_name, Service_code, Agency_responsible, Address, Channel, Media_url string
	Requested_datetime, Updated_datetime                                                                           string // FIXME: should these be proper time objects?
	Extended_attributes     map[string]interface{}
}

func main() {
	db, err := sql.Open("postgres", "dbname=cwfy sslmode=disable")
	if err == nil {
		requests := fetchRequests()
		for _, request := range requests {

			if request.Service_request_id == "" {
				log.Printf("Ignoring a request type %s because there is no SR number assigned", request.Service_name)
			} else {
				stmt, err := db.Prepare("INSERT INTO service_requests(service_request_id," +
					"status, service_name, service_code, agency_responsible, " +
					"address, requested_datetime, updated_datetime, lat, long," +
					"ward, police_district, media_url, channel, duplicate, parent_service_request_id) " +
					"VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16);")

				if err != nil {
					log.Fatal("error preparing database statement", err)
				}

                                _, err = stmt.Exec(request.Service_request_id,
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
                                     log.Printf("saved SR %s to database", request.Service_request_id)
                                }
			}
		}
	} else {
		log.Fatal("Cannot open database connection", err)
	}
}

func (req Open311Request) String() string {
	return fmt.Sprintf("%s: %s at %s %f,%f", req.Service_request_id, req.Service_name, req.Address, req.Lat, req.Long)
}

func fetchRequests() (requests []Open311Request) {
	log.Printf("fetching from %s", OPEN311_API_URI)
	resp, err := http.Get(OPEN311_API_URI)
	defer resp.Body.Close()

	if err == nil {
		log.Println("fetch succesful, reading response")
		body, err := ioutil.ReadAll(resp.Body)

		if err == nil {
			log.Println("loaded response body.")
			err := json.Unmarshal(body, &requests)
			if err != nil {
				log.Fatal("error parsing JSON:", err)
			}

			log.Printf("received %d requests from Open311", len(requests))

		}
	} else {
		log.Fatalln("error fetching from Open311 endpoint", err)
	}

	return requests
}
