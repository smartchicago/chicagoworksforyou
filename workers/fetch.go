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
	service_request_id, status, service_name, service_code, agency_responsible, address string
	lat, long                                                                           float32
	requested_datetime, updated_datetime                                                string // FIXME: should these be proper time objects?
}

func main() {
	fetchRequests()
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
			var requests []map[string]interface{}
			err := json.Unmarshal(body, &requests)
			if err != nil {
				log.Fatal("error parsing JSON:", err)
			}
			
			for _, req := range requests {
				fmt.Println(req)
			}
		}
	} else {
		log.Fatalln("error fetching from Open311 endpoint", err)
	}
}
