package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type HighDay struct {
	Date  string
	Count int
}

func WardHistoricHighsHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
	// given a ward and service type, return the set of days with the most SR opened
	//
	// Parameters:
	// 	count: 		number of historicl high days to return.
	//	service_code:   (optional) the code used by the City of Chicago to categorize service requests. If omitted, all services codes will be returned
	//	callback:       function to wrap response in (for JSONP functionality)
	// 	include_date:  	pass a YYYY-MM-DD string and the count for that day will be included
	//

	vars := mux.Vars(request)
	ward_id := vars["id"]

	days, err := strconv.Atoi(params.Get("count"))
	if err != nil || days < 1 || days > 60 {
		return nil, &ApiError{Msg: "invalid count, must be integer, 1..60", Code: 400}
	}

	service_code = params.Get("service_code")

	chi, _ := time.LoadLocation("America/Chicago")
	day, err := time.ParseInLocation("2006-01-02", params.Get("include_date"), chi)
	if err != nil {
		return nil, &ApiError{Msg: "invalid include_date", Code: 400}
	}

	// if service_code provided, find highs for that code
	// otherwise, find highs for each service code

	if service_code != "" {
		counts := findAllTimeHighs(service_code, ward_id, days)

		if !day.IsZero() {
			counts = append(counts, HighDay{Date: day.Format("2006-01-02"), Count: findDayTotal(service_code, ward_id, day)})
		}

		return dumpJson(counts), nil

	} else {
		type ResponseData struct {
			Highs   map[string][]HighDay
			Current map[string]HighDay
		}

		var resp ResponseData
		resp.Highs = make(map[string][]HighDay)
		resp.Current = make(map[string]HighDay)

		for _, code := range ServiceCodes {
			// find highs
			resp.Highs[code] = findAllTimeHighs(code, ward_id, days)

			// find date, if spec'd
			if !day.IsZero() {
				resp.Current[code] = HighDay{Date: day.Format("2006-01-02"), Count: findDayTotal(code, ward_id, day)}
			}
		}

		return dumpJson(resp), nil
	}

}
func findAllTimeHighs(service_code string, ward_id string, days int) (counts []HighDay) {
	rows, err := api.Db.Query(`SELECT total,requested_date
		FROM daily_counts
		WHERE service_code = $1
			AND ward = $2
		ORDER BY total DESC, requested_date DESC
		LIMIT $3;`, service_code, ward_id, days)

	if err != nil {
		log.Print("error fetching historic highs ", err)
	}

	for rows.Next() {
		var d time.Time
		var dc HighDay

		if err := rows.Scan(&dc.Count, &d); err != nil {
			log.Print("error loading high value ", err)
		}
		counts = append(counts, HighDay{Date: d.Format("2006-01-02"), Count: dc.Count})
	}

	return
}

func findDayTotal(service_code string, ward_id string, day time.Time) (count int) {
	err := api.Db.QueryRow(`SELECT total
		FROM daily_counts
		WHERE service_code = $1
			AND ward = $2
			AND requested_date = $3;
		`, service_code, ward_id, day).Scan(&count)

	if err != nil {
		// no rows
		count = 0
	}

	return
}
