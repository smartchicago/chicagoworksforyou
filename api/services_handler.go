package main

import (
	"net/http"
	"net/url"
)

var ServiceNames = map[string]string{"4fd3b167e750846744000005": "Graffiti Removal",
	"4fd6e4ece750840569000019": "Restaurant Complaint",
	"4fd3b9bce750846c5300004a": "Rodent Baiting / Rat Complaint",
	"4fd3bbf8e750846c53000069": "Tree Debris",
	"4ffa4c69601827691b000018": "Abandoned Vehicle",
	"4ffa9f2d6018277d400000c8": "Street Light 1 / Out",
	"4ffa971e6018277d4000000b": "Pavement Cave-In Survey",
	"4ffa9cad6018277d4000007b": "Alley Light Out",
	"4fd3bd72e750846c530000cd": "Building Violation",
	"4ffa9db16018277d400000a2": "Traffic Signal Out",
	"4ffa995a6018277d4000003c": "Street Cut Complaints",
	"4fd3b750e750846c5300001d": "Sanitation Code Violation",
	"4fd3b656e750846c53000004": "Pothole in Street",
	"4fd3bd3de750846c530000b9": "Street Lights All / Out"}

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

	rows, err := api.Db.Query(`SELECT SUM(total), service_code
	        FROM daily_counts
	        GROUP BY service_code;`)

	if err != nil {
		return backend_error(err)
	}

	for rows.Next() {
		var count int
		var service_code string

		if err := rows.Scan(&count, &service_code); err != nil {
			return backend_error(err)
		}

		row := ServicesCount{Count: count, Service_code: service_code, Service_name: ServiceNames[service_code]}
		services = append(services, row)
	}

	return dumpJson(services), nil
}
