package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/bmizerany/pq"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	log.Print("starting ChicagoWorksforYou.com API server")

	router := mux.NewRouter()
	router.HandleFunc("/health_check", HealthCheckHandler)
	router.HandleFunc("/services.json", ServicesHandler)
	http.ListenAndServe(":4000", router)
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
	health_check := map[string]string{}
	health_check["database"] = "dbconn" // FIXME: meaningful db information
	health_check["sr_count"] = "123"    // FIXME: meaningful count
	jsn, _ := json.Marshal(health_check)
	response.Write(jsn)
}
