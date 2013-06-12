package main

import (
	"database/sql"
	"encoding/json"
	"github.com/bmizerany/pq"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	log.Print("starting ChicagoWorksforYou.com API server")

	router := mux.NewRouter()
	router.HandleFunc("/health_check", HealthCheckHandler)
	router.HandleFunc("/services.json", ServicesHandler)
	router.HandleFunc("/wards/{id}/requests.json", WardRequestsHandler)
	http.ListenAndServe(":4000", router)
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
		Lat, Long                                                                                               float64
		Ward, Police_district, Id                                                                                   int
		Service_request_id, Status, Service_name, Service_code, Agency_responsible, Address, Channel, Media_url, Duplicate, Parent_service_request_id sql.NullString
		Requested_datetime, Updated_datetime, Created_at, Updated_at                                                                    pq.NullTime // FIXME: should these be proper time objects?
		Extended_attributes                                                                                     map[string]interface{}
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
		Count int
		Database bool
		Healthy bool
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
