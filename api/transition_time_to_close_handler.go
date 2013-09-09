package main

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func TransitionTimeToCloseHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
	// Given transition zone id, service type, date, length of time & increment,
	// return time-to-close for that service type, for each
	// increment over that length of time, going backwards from that date.
	// If service_code is omitted, the average will be for all service types.
	//
	// Response data:
	//      "Count" is the number of service requests closed in the given time period.
	//      "Time" is the average difference, in days, between closed and requested datetimes.
	//
	// Sample request and output:
	//
	// 	$ curl "http://localhost:5000/transitions/time_to_close.json?transition_area_id=1&count=7&end_date=2013-08-22"
	//	 {
	//   		"Time": 0.04724537037037037,
	//   		"Count": 1
	// 	}

	// required
	transition_area_id, err := strconv.Atoi(params.Get("transition_area_id"))
	if transition_area_id == 0 || err != nil {
		return nil, &ApiError{Code: 400, Msg: "transition_area_id is required and must be an integer"}
	}

	service_code := params.Get("service_code")
	days, err := strconv.Atoi(params.Get("count"))
	if err != nil || days < 1 || days > 60 {
		return nil, &ApiError{Msg: "invalid count, must be integer, 1..60", Code: 400}
	}

	chi, _ := time.LoadLocation("America/Chicago")
	end, err := time.ParseInLocation("2006-01-02", params.Get("end_date"), chi)
	if err != nil {
		return nil, &ApiError{Msg: "invalid end_date", Code: 400}
	}

	end = end.AddDate(0, 0, 1) // inc to the following day
	start := end.AddDate(0, 0, -days)

	type TimeToClose struct {
		Time  float64
		Count int
	}
	var ttc TimeToClose

	if service_code != "" {
		log.Printf("fetching a single service code %s", service_code)
		err = api.Db.QueryRow(`SELECT EXTRACT('EPOCH' FROM AVG(closed_datetime - requested_datetime)) AS avg_ttc, COUNT(service_request_id)
        		FROM service_requests 
        		WHERE closed_datetime IS NOT NULL 
        			AND duplicate IS NULL
        			AND closed_datetime >= $1
        			AND closed_datetime <= $2
        			AND service_code = $3 
				AND transition_area_id = $4
        		ORDER BY avg_ttc DESC;`, start, end, service_code, transition_area_id).Scan(&ttc.Time, &ttc.Count)
	} else {
		log.Printf("fetching all service codes")
		err = api.Db.QueryRow(`SELECT EXTRACT('EPOCH' FROM AVG(closed_datetime - requested_datetime)) AS avg_ttc, COUNT(service_request_id)
        		FROM service_requests 
        		WHERE closed_datetime IS NOT NULL 
        			AND duplicate IS NULL
        			AND closed_datetime >= $1
        			AND closed_datetime <= $2
        			AND transition_area_id = $3
        		ORDER BY avg_ttc DESC;`, start, end, transition_area_id).Scan(&ttc.Time, &ttc.Count)
	}

	if err != nil {
		log.Print("error fetching time to close", err)
	}

	ttc.Time = ttc.Time / 86400.0 // convert from seconds to days

	return dumpJson(ttc), nil
}
