package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func WardCountsHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
	// for a given ward, return the number of service requests opened
	// grouped by day, then by service request type
	//
	// Parameters:
	//
	//	count:          the number of days of data to return
	//	end_date:       date that +count+ is based from.
	//	service_code:   the code used by the City of Chicago to categorize service requests
	//	callback:       function to wrap response in (for JSONP functionality)
	//
	// Sample API output
	//
	// Note that the end date is June 12, and the results include the end_date. Days with no service requests will report "0"
	//
	// $ curl "http://localhost:5000/wards/10/counts.json?service_code=4fd3b167e750846744000005&count=7&end_date=2013-07-03"
	// {
	//   "2013-06-27": {
	//     "Count": 4,
	//     "CityTotal": 440,
	//     "CityAverage": 8.8
	//   },
	//   "2013-06-28": {
	//     "Count": 8,
	//     "CityTotal": 372,
	//     "CityAverage": 7.44
	//   },
	//   "2013-06-29": {
	//     "Count": 1,
	//     "CityTotal": 93,
	//     "CityAverage": 1.86
	//   },

	vars := mux.Vars(request)
	ward_id := vars["id"]

	// determine date range.
	days, _ := strconv.Atoi(params["count"][0])

	chi, _ := time.LoadLocation("America/Chicago")
	end, _ := time.ParseInLocation("2006-01-02", params["end_date"][0], chi)
	end = end.AddDate(0, 0, 1) // inc to the following day
	start := end.AddDate(0, 0, -days)

	service_code := params["service_code"][0]

	rows, err := api.Db.Query(`SELECT requested_date, SUM(dc.total) AS opened, SUM(dcc.total) AS closed
		FROM daily_counts dc
		INNER JOIN daily_closed_counts dcc
		USING(requested_date, ward, service_code)
		WHERE ward = $1
			AND service_code = $2
			AND requested_date >= $3
			AND requested_date <= $4
		GROUP BY requested_date
		ORDER BY requested_date DESC;`, ward_id, service_code, start, end)

	if err != nil {
		log.Fatal("error fetching data for WardCountsHandler", err)
	}

	type WardCount struct {
		Opened      int
		Closed      int
		CityTotal   int
		CityAverage float32
	}

	counts := make(map[string]WardCount)
	for rows.Next() {
		var wc WardCount
		var rd time.Time

		if err := rows.Scan(&rd, &wc.Opened, &wc.Closed); err != nil {
			log.Print("error reading row of ward count", err)
		}

		counts[rd.Format("2006-01-02")] = wc
	}

	// calculate the citywide average for each day
	rows, err = api.Db.Query(`SELECT COUNT(*), DATE(requested_datetime) AS requested_date 
		FROM service_requests 
		WHERE duplicate IS NULL 
			AND service_code = $1 
			AND requested_datetime >= $2::date 
			AND requested_datetime <= $3::date
		GROUP BY DATE(requested_datetime) 
		ORDER BY requested_date;`,
		service_code, start, end)

	if err != nil {
		log.Fatal("error fetching data for WardCountsHandler", err)
	}

	for rows.Next() {
		var rd time.Time
		var city_total int
		if err := rows.Scan(&city_total, &rd); err != nil {
			log.Print("error reading row of ward count", err)
		}

		k := rd.Format("2006-01-02")
		tmp := counts[k]
		tmp.CityTotal = city_total
		tmp.CityAverage = float32(city_total) / 50.0
		counts[k] = tmp
	}

	resp := make(map[string]WardCount)
	for i := 1; i < days+1; i++ { // note: we inc. end to the following day above, so need to compensate here otherwise it's off-by-one
		d := end.AddDate(0, 0, -i)
		key := d.Format("2006-01-02")
		resp[key] = counts[key]
	}

	return dumpJson(resp), nil
}
