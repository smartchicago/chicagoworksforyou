package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const OPEN311_API_URI = "http://311api.cityofchicago.org/open311/v2/requests.json"

type Open311Request struct {
	Lat, Long                                                                           float64
	Service_request_id, Status, Service_name, Service_code, Agency_responsible, Address string
	Requested_datetime, Updated_datetime                                                string // FIXME: should these be proper time objects?
}

func main() {
	fetchRequests()
}

func (req Open311Request) String() string {
	return fmt.Sprintf("%s: %s at %s %f,%f", req.Service_request_id, req.Service_name, req.Address, req.Lat, req.Long)
}

func fetchRequests() {
	log.Printf("fetching from %s", OPEN311_API_URI)
	resp, err := http.Get(OPEN311_API_URI)
	defer resp.Body.Close()

	if err == nil {
		log.Println("fetch succesful, reading response")
		body, err := ioutil.ReadAll(resp.Body)

		if err == nil {
			log.Println("loaded response body.")
			// parse into JSON

			var requests []Open311Request
			err := json.Unmarshal(body, &requests)
			if err != nil {
				log.Fatal("error parsing JSON:", err)
			}

			log.Printf("received %d requests from Open311", len(requests))
			
			for i, req := range requests {
				fmt.Println(i, req)
			}
		}
	} else {
		log.Fatalln("error fetching from Open311 endpoint", err)
	}
}
