package main

import (
	"log"
	"net/http"
	"net/url"
)

func HealthCheckHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
	type HealthCheck struct {
		SR       string `json:"most_recent_sr_id"`
		Database bool   `json:"database"`
		Healthy  bool   `json:"healthy"`
		Version  string `json:"version"`
	}

	health_check := HealthCheck{Version: version}
	health_check.Database = api.Db.Ping() == nil

	err := api.Db.QueryRow("SELECT service_request_id FROM service_requests ORDER BY requested_datetime DESC LIMIT 1").Scan(&health_check.SR)
	if err != nil {
		return backend_error(err)
	}

	// calculate overall health
	health_check.Healthy = health_check.SR != "" && health_check.Database

	log.Printf("health_check: %+v", health_check)
	if !health_check.Healthy {
		log.Printf("health_check failed")
	}

	return dumpJson(health_check), nil
}
