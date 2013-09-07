package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func WardCountsHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
	// for a given ward, return the number of service requests opened and closed
	// grouped by day, then by service request type
	//
	// Parameters:
	//
	//	count:          the number of days of data to return
	//	end_date:       date that +count+ is based from.
	//	service_code:   (optional) the code used by the City of Chicago to categorize service requests
	//	callback:       function to wrap response in (for JSONP functionality)
	//
	// Sample API output
	//
	// Note that the end date is August 30, and the results include the end_date. Days with no service requests will report "0"
	//
	// $ curl "http://localhost:5000/wards/10/counts.json?count=7&end_date=2013-08-30"
	// {
	//   "2013-08-24": {
	//     "Opened": 0,
	//     "Closed": 0
	//   },
	//   "2013-08-25": {
	//     "Opened": 0,
	//     "Closed": 0
	//   },
	//   "2013-08-26": {
	//     "Opened": 7,
	//     "Closed": 4
	//   },
	//   "2013-08-27": {
	//     "Opened": 20,
	//     "Closed": 37
	//   },
	//   "2013-08-28": {
	//     "Opened": 18,
	//     "Closed": 34
	//   },
	//   "2013-08-29": {
	//     "Opened": 7,
	//     "Closed": 6
	//   },
	//   "2013-08-30": {
	//     "Opened": 0,
	//     "Closed": 0
	//   }
	// }

	vars := mux.Vars(request)
	ward_id := vars["id"]

	// determine date range.

	days, _ := strconv.Atoi(params.Get("count"))

	chi, _ := time.LoadLocation("America/Chicago")
	end, err := time.ParseInLocation("2006-01-02", params.Get("end_date"), chi)
	if err != nil {
		return nil, &ApiError{Msg: "invalid end_date", Code: 400}
	}

	end = end.AddDate(0, 0, 1) // inc to the following day
	start := end.AddDate(0, 0, -days)

	service_code := params.Get("service_code")

	query := `SELECT requested_date, SUM(dc.total) AS opened, SUM(dcc.total) AS closed
		FROM daily_counts dc
		INNER JOIN daily_closed_counts dcc
		USING(requested_date, ward, service_code)
		WHERE ward = $1
			AND requested_date >= $2
			AND requested_date <= $3
			%s
		GROUP BY requested_date
		ORDER BY requested_date DESC;`

	var rows *sql.Rows
	var err error

	if service_code != "" {
		query = fmt.Sprintf(query, "AND service_code = $4")
		rows, err = api.Db.Query(query, ward_id, start, end, service_code)
	} else {
		query = fmt.Sprintf(query, "")
		rows, err = api.Db.Query(query, ward_id, start, end)
	}

	if err != nil {
		log.Fatal("error fetching data for WardCountsHandler", err)
	}

	type WardCount struct {
		Opened int
		Closed int
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

	resp := make(map[string]WardCount)
	for i := 1; i < days+1; i++ { // note: we inc. end to the following day above, so need to compensate here otherwise it's off-by-one
		d := end.AddDate(0, 0, -i)
		key := d.Format("2006-01-02")
		resp[key] = counts[key]
	}

	return dumpJson(resp), nil
}
