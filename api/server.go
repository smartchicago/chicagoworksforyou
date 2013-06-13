package main

import (
	"database/sql"
	"encoding/json"
	"github.com/bmizerany/pq"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

func main() {
	log.Print("starting ChicagoWorksforYou.com API server")

	router := mux.NewRouter()
	router.HandleFunc("/health_check", HealthCheckHandler)
	router.HandleFunc("/services.json", ServicesHandler)
	router.HandleFunc("/wards/{id}/requests.json", WardRequestsHandler)
	router.HandleFunc("/wards/{id}/counts.json", WardCountsHandler)
	http.ListenAndServe(":5000", router)
}

func WardCountsHandler(response http.ResponseWriter, request *http.Request) {
	// for a given ward, return the number of service requests opened
	// grouped by day, then by service request type

        // sample output
        // $ curl "http://localhost:5000/wards/10/counts.json?service_code=4fd3b167e750846744000005"
        // {
        //   "2013-06-06": 2,
        //   "2013-06-07": 4,
        //   "2013-06-09": 5,
        //   "2013-06-10": 6,
        //   "2013-06-12": 23
        // }
        //
	vars := mux.Vars(request)
	ward_id := vars["id"]
	params := request.URL.Query()

	db, err := sql.Open("postgres", "dbname=cwfy sslmode=disable")
	if err != nil {
		log.Fatal("Cannot open database connection", err)
	}
	defer db.Close()

	log.Printf("fetching counts for ward %s code %s", ward_id, params["service_code"][0])

	rows, err := db.Query("SELECT COUNT(*), DATE(requested_datetime) as requested_date FROM service_requests WHERE ward = $1 AND duplicate IS NULL AND service_code = $2 GROUP BY DATE(requested_datetime) ORDER BY requested_date;", string(ward_id), params["service_code"][0])
	if err != nil {
		log.Fatal("error fetching data for WardCountsHandler", err)
	}

	type WardCount struct {
		Requested_date time.Time
		Count          int
	}

	var counts []WardCount
	for rows.Next() {
		wc := WardCount{}
		if err := rows.Scan(&wc.Count, &wc.Requested_date); err != nil {
			log.Print("error reading row of ward count", err)
		}

		// trunc the requested time to just date
		counts = append(counts, wc)
	}

	resp := make(map[string]int)

	for _, c := range counts {
		key := c.Requested_date.Format("2006-01-02")
		log.Print("key: ", key)
		resp[key] = c.Count
	}

	jsn, _ := json.MarshalIndent(resp, "", "  ")
	response.Write(jsn)
}

func WardRequestsHandler(response http.ResponseWriter, request *http.Request) {
	// for a given ward, return recent service requests

	vars := mux.Vars(request)
	ward_id := vars["id"]

	log.Print("fetch requests for ward ", ward_id)

	// open database
	db, err := sql.Open("postgres", "dbname=cwfy sslmode=disable")
	if err != nil {
		log.Fatal("Cannot open database connection", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT lat,long,ward,police_district,service_request_id,status,service_name,service_code,agency_responsible,address,channel,media_url,requested_datetime,updated_datetime,created_at,updated_at,duplicate,parent_service_request_id,id FROM service_requests WHERE duplicate IS NULL AND ward = $1 ORDER BY updated_at DESC LIMIT 100;", ward_id)

	if err != nil {
		log.Fatal("error fetching data for WardRequestsHandler", err)
	}

	type Open311RequestRow struct {
		Lat, Long                                                                                                                                     float64
		Ward, Police_district, Id                                                                                                                     int
		Service_request_id, Status, Service_name, Service_code, Agency_responsible, Address, Channel, Media_url, Duplicate, Parent_service_request_id sql.NullString
		Requested_datetime, Updated_datetime, Created_at, Updated_at                                                                                  pq.NullTime // FIXME: should these be proper time objects?
		Extended_attributes                                                                                                                           map[string]interface{}
	}

	var result []Open311RequestRow

	for rows.Next() {
		var row Open311RequestRow
		if err := rows.Scan(&row.Lat, &row.Long, &row.Ward, &row.Police_district,
			&row.Service_request_id, &row.Status, &row.Service_name,
			&row.Service_code, &row.Agency_responsible, &row.Address,
			&row.Channel, &row.Media_url, &row.Requested_datetime,
			&row.Updated_datetime, &row.Created_at, &row.Updated_at,
			&row.Duplicate, &row.Parent_service_request_id,
			&row.Id); err != nil {
			log.Fatal("error reading row", err)
		}

		result = append(result, row)
	}

	jsn, _ := json.MarshalIndent(result, "", "  ")
	response.Write(jsn)
}

func ServicesHandler(response http.ResponseWriter, request *http.Request) {
	// return counts of requests, grouped by service name
	//
	// Sample output:
	//
	// [
	//   {
	//     "Count": 1139,
	//     "Service_code": "4fd3b167e750846744000005",
	//     "Service_name": "Graffiti Removal"
	//   },
	//   {
	//     "Count": 25,
	//     "Service_code": "4fd6e4ece750840569000019",
	//     "Service_name": "Restaurant Complaint"
	//   },
	//
	//  ... snip ...
	//
	// ]

	type ServicesCount struct {
		Count        int
		Service_code string
		Service_name string
	}

	var services []ServicesCount

	// open database
	db, err := sql.Open("postgres", "dbname=cwfy sslmode=disable")
	if err != nil {
		log.Fatal("Cannot open database connection", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT COUNT(*), service_code, service_name FROM service_requests WHERE duplicate IS NULL GROUP BY service_code,service_name;")

	if err != nil {
		log.Fatal("error fetching data for ServicesHandler", err)
	}

	for rows.Next() {
		var count int
		var service_code, service_name string

		if err := rows.Scan(&count, &service_code, &service_name); err != nil {
			log.Fatal("error reading row", err)
		}

		row := ServicesCount{Count: count, Service_code: service_code, Service_name: service_name}
		services = append(services, row)
	}

	jsn, _ := json.MarshalIndent(services, "", "  ")
	response.Write(jsn)
}

func HealthCheckHandler(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("Content-type", "application/json")

	type HealthCheck struct {
		Count    int
		Database bool
		Healthy  bool
	}

	// open database
	db, err := sql.Open("postgres", "dbname=cwfy sslmode=disable")
	if err != nil {
		log.Fatal("Cannot open database connection", err)
	}
	defer db.Close()

	health_check := HealthCheck{}

	health_check.Database = db.Ping() == nil

	rows, _ := db.Query("SELECT COUNT(*) FROM service_requests;")
	for rows.Next() {
		if err := rows.Scan(&health_check.Count); err != nil {
			log.Fatal("error fetching count", err)
		}
	}

	// calculate overall health
	health_check.Healthy = health_check.Count > 0 && health_check.Database

	log.Printf("health_check: %+v", health_check)
	if !health_check.Healthy {
		log.Printf("health_check failed")
	}
	jsn, _ := json.Marshal(health_check)
	response.Write(jsn)
}
