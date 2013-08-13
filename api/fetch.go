package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"
	"io/ioutil"

)

const OPEN311_API_URI = "http://311api.cityofchicago.org/open311/v2/requests.json?extensions=true&page_size=500"

type Worker struct {
	// Db         *sql.DB
	LastRunAt time.Time
}

var worker Worker
var srdb ServiceRequestDB
var env Environment

//  command line flags
var (
	version       string // set at compile time, will be the current git hash
	environment   = flag.String("environment", "", "Environment to run in, e.g. staging, production")
	config        = flag.String("config", "./config/database.yml", "database configuration file")
	// sr_number     = flag.String("sr-number", "", "SR number to fetch")
	backfill      = flag.Bool("backfill", false, "run in reverse and backfill data")
	backfill_date = flag.String("backfill-from", time.Now().Format(time.RFC3339), "date to start backfilling data from. Use RFC3339 format. Default will be the current time.")
)

func init() {
	flag.Parse()

	log.Printf("CWFY Fetch worker version %s running in %s environment, configuration file %s", version, *environment, *config)

	srdb.Init(env.Load(config, environment))
	// todo: error handling
}

func main() {
	defer srdb.Close()

	// if *sr_number != "" {
	// 	sr := fetchSingleRequest(*sr_number)
	// 	srdb.Save(sr)
	// 	return
	// }

	start_backfill_from := *backfill_date
	for {
		switch {
		case *backfill:
			requests := backFillRequests(start_backfill_from)
			for _, request := range requests {
				srdb.Save(&request)
			}

			start_backfill_from = requests[len(requests)-1].Updated_datetime.Format(time.RFC3339)

		case time.Since(worker.LastRunAt) > (30 * time.Second):
			// load requests from open311
			for _, request := range fetchRequests() {
				srdb.Save(&request)
			}
			worker.LastRunAt = time.Now()
		default:
			log.Print("sleeping for 10 seconds")
			time.Sleep(10 * time.Second)
		}
	}
}

// func fetchSingleRequest(sr_number string) (request ServiceRequest) {
// 	// given an SR, fetch the record
// 	log.Printf("fetching single SR %s", sr_number)
// 	open311_api_endpoint := fmt.Sprintf("http://311api.cityofchicago.org/open311/v2/requests/%s.json?extensions=true", sr_number)
// 
// 	log.Printf("fetching from %s", open311_api_endpoint)
// 	resp, err := http.Get(open311_api_endpoint)
// 	if err != nil {
// 		log.Fatal("error fetching from Open311 endpoint", err)
// 	}
// 	defer resp.Body.Close()
// 
// 	// load response body
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Fatal("error loading response body", err)
// 	}
// 
// 	// parse JSON and load into an array of ServiceRequest objects
// 	var requests []ServiceRequest
// 
// 	err = json.Unmarshal(body, &requests)
// 	if err != nil {
// 		log.Fatal("error parsing JSON:", err)
// 	}
// 
// 	log.Printf("received %d requests from Open311", len(requests))
// 
// 	return requests[0]
// }

func fetchRequests() (requests []ServiceRequest) {
	last_updated_at := time.Now()

	newest, _ := srdb.Newest()	// FIXME: error handling
	if newest != nil {
		// override with the most recent SR available
		last_updated_at = newest.Updated_datetime
	}

	log.Print("[fetchRequests] most recent SR timestamp ", last_updated_at.Format(time.RFC3339))

	// construct the request URI using base params and the proper time
	open311_api_endpoint := OPEN311_API_URI + "&updated_after=" + last_updated_at.Format(time.RFC3339)

	log.Printf("[fetchRequests] fetching from %s", open311_api_endpoint)

	http.DefaultTransport.(*http.Transport).ResponseHeaderTimeout = time.Second * 60

	resp, err := http.Get(open311_api_endpoint)

	if err != nil {
		log.Fatalln("[fetchRequests] error fetching from Open311 endpoint", err)
	}

	defer resp.Body.Close()

	// load response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("[fetchRequests] error loading response body", err)
	}

	// parse JSON and load into an array of ServiceRequest objects
	err = json.Unmarshal(body, &requests)
	if err != nil {
		log.Fatal("[fetchRequests] error parsing JSON:", err)
	}

	log.Printf("[fetchRequests] received %d requests from Open311", len(requests))

	return requests
}

func backFillRequests(start_from string) (requests []ServiceRequest) {
	var fetch_from time.Time

	if start_from == "" {
		oldest, _ := srdb.Oldest() // FIXME: error handling
		fetch_from = oldest.Updated_datetime
		log.Printf("no start_from value provided, so falling back to oldest (by last update) SR in the database: %s", fetch_from)
	} else {
		t, err := time.Parse(time.RFC3339, start_from)
		if err != nil {
			log.Fatal("[backfill] error parsing date to start from", err)
		}
		fetch_from = t
	}

	formatted_date_string_with_tz := fetch_from.Format(time.RFC3339)

	// construct the request URI using base params and the proper time
	open311_api_endpoint := OPEN311_API_URI + "&updated_before=" + formatted_date_string_with_tz

	log.Printf("[backfill] fetching from %s", open311_api_endpoint)
	http.DefaultTransport.(*http.Transport).ResponseHeaderTimeout = time.Second * 60

	resp, err := http.Get(open311_api_endpoint)
	if err != nil {
		log.Fatalln("[backfill] error fetching from Open311 endpoint", err)
	}
	defer resp.Body.Close()

	// load response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("[backfill] error loading response body", err)
	}

	// parse JSON and load into an array of ServiceRequest objects
	err = json.Unmarshal(body, &requests)
	if err != nil {
		log.Fatal("[backfill] error parsing JSON:", err)
	}

	log.Printf("[backfill] received %d requests from Open311", len(requests))

	return requests
}
