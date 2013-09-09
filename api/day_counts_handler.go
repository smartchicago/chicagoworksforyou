package main

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func DayCountsHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
	// Given day, return total # of each service type,
	// along with daily average for each service type and
	// wards that opened the most requests that day.
	//
	// $ curl "http://localhost:5000/requests/counts_by_day.json?day=2013-06-21"
	//         {
	//           "4fd3b167e750846744000005": {
	//             "Count": 379,
	//             "Average": 8.694054,
	//             "TopWards": [
	//               14
	//             ]
	//           },
	//           "4fd3b9bce750846c5300004a": {
	//             "Count": 86,
	//             "Average": 2.774941,
	//             "TopWards": [
	//               32,
	//               50
	//             ]
	//           },

	chi, _ := time.LoadLocation("America/Chicago")
	end, err := time.ParseInLocation("2006-01-02", params.Get("day"), chi)
	if err != nil {
		return nil, &ApiError{Msg: "invalid date", Code: 400}
	}

	end = end.AddDate(0, 0, 1) // inc to the following day
	start := end.AddDate(0, 0, -1)

	type DayCount struct {
		Count   int
		Average float32
		Wards   map[string]int
	}

	counts := make(map[string]DayCount)

	// fetch the total number of SR opened by service code

	rows, err := api.Db.Query(`SELECT service_code, SUM(total) AS cnt 
             FROM daily_counts
             WHERE requested_date >= $1 
                     AND requested_date < $2
             GROUP BY service_code
             ORDER BY cnt;`, start, end)

	if err != nil {
		return backend_error(err)
	}

	for rows.Next() {
		var dc DayCount
		var sc string
		if err := rows.Scan(&sc, &dc.Count); err != nil {
			return backend_error(err)
		}
		counts[sc] = dc
	}

	// fetch top ward(s) for each service_code
	// for each service code, find the ward (or wards) with the most SR opened

	// for each service code, fetch wards for day sorted by # SR opened

	for _, sc := range ServiceCodes {
		wards := make(map[int]int) // map the ward to its total number of reqs for the service code for the day

		rows, err := api.Db.Query(`SELECT total, ward 
                     FROM daily_counts
                     WHERE requested_date >= $1 
                             AND requested_date < $2
                             AND service_code = $3
                     ORDER BY total DESC;`, start, end, sc)

		if err != nil {
			return backend_error(err)
		}

		for rows.Next() {
			var ward, total int
			if err := rows.Scan(&total, &ward); err != nil {
				return backend_error(err)
			}
			wards[ward] = total
		}

		// zero fill wards that are not present, append to response struct
		tmp := counts[sc]
		tmp.Wards = make(map[string]int)
		for i := 1; i < 51; i++ {
			if _, present := wards[i]; !present {
				wards[i] = 0
			}

			tmp.Wards[strconv.Itoa(i)] = wards[i]
		}
		counts[sc] = tmp

	}

	// fetch daily averages
	rows, err = api.Db.Query(`SELECT service_code, SUM(total)/365.0 AS avg_reports 
		FROM daily_counts 
		WHERE requested_date >= (NOW() - INTERVAL '1 year')
		GROUP BY service_code
		ORDER BY avg_reports;`)

	if err != nil {
		return backend_error(err)
	}

	for rows.Next() {
		var sc string
		var avg float32
		if err := rows.Scan(&sc, &avg); err != nil {
			return backend_error(err)
		}

		tmp := counts[sc]
		tmp.Average = avg
		counts[sc] = tmp
	}

	return dumpJson(counts), nil
}
