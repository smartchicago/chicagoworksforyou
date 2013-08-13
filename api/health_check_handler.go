package main

import (
	"log"
	"net/http"
	"net/url"
)

func HealthCheckHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
	type HealthCheck struct {
		Count             int
		Database, Healthy bool
		Version           string
	}

	health_check := HealthCheck{Version: version}
	health_check.Database = api.Db.Ping() == nil

	rows, _ := api.Db.Query("SELECT COUNT(*) FROM service_requests;")
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

	return dumpJson(health_check), nil
}
