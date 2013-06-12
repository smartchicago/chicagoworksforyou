package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/bmizerany/pq"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var db *sql.DB

func main() {
	log.Print("starting ChicagoWorksforYou.com API server")

	// open database
	db, err := sql.Open("postgres", "dbname=cwfy sslmode=disable")
	if err != nil {
		log.Fatal("Cannot open database connection", err)
	}
	defer db.Close()

        log.Print("db is", db)

	router := mux.NewRouter()
	router.HandleFunc("/health_check", HealthCheckHandler)
	router.HandleFunc("/services.json", ServicesHandler)
	http.ListenAndServe(":4000", router)
}

// type ServicesCount struct {
//      Count        int
//      Service_code string
//      Service_name string
// }

func ServicesHandler(response http.ResponseWriter, request *http.Request) {
	// return counts of requests, grouped by service name

        // var services []ServicesCount

        log.Print("db is now", db)

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
		log.Print(count, service_code, service_name)
	}
	
        // jsn, _ := json.Marshal(services)
        // response.Write()
}

func HealthCheckHandler(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("Content-type", "application/json")
	health_check := map[string]string{}
	health_check["database"] = "dbconn" // FIXME: meaningful db information
	health_check["sr_count"] = "123"    // FIXME: meaningful count
	jsn, _ := json.Marshal(health_check)
	response.Write(jsn)
}
