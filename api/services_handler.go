package main

import (
	"log"
	"net/http"
	"net/url"
)

func ServicesHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
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

	rows, err := api.Db.Query("SELECT COUNT(*), service_code, service_name FROM service_requests WHERE duplicate IS NULL GROUP BY service_code,service_name;")

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

	return dumpJson(services), nil
}
