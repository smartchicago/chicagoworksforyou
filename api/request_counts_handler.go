package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func RequestCountsHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
	// for a given request service type and date, return the count
	// of requests for that date, grouped by ward, and the city total
	// The output is a map where keys are ward identifiers, and the value is the count.
	//
	// Sample request and output:
	// $ curl "http://localhost:5000/requests/4fd3b167e750846744000005/counts.json?end_date=2013-06-10&count=1"
	// {
	//   "DayData": [
	//     "2013-06-10"
	//   ],
	//   "CityData": {
	//     "Average": 1.5424658,
	//     "DailyMax": [
	//       114,
	//       106,
	//       104,
	//       102,
	//       102,
	//       94,
	//       93
	//     ],
	//     "Count": 563
	//   },
	//   "WardData": {
	//     "1": {
	//       "Counts": [
	//         33
	//       ],
	//       "Average": 6.917808
	//     },
	//     "10": {
	//       "Counts": [
	//         6
	//       ],
	//       "Average": 2.4958904
	//     },
	//     "11": {
	//       "Counts": [
	//         26
	//       ],
	//       "Average": 8.087671
	//     },
	//     (... truncated ...)
	//   }
	// }

	vars := mux.Vars(request)
	service_code := vars["service_code"]

	// determine date range.
	days, err := strconv.Atoi(params.Get("count"))
	if err != nil || days > 60 || days < 1 {
		return nil, &ApiError{Msg: "invalid count, must be integer, 1..60", Code: 400}
	}

	chi, _ := time.LoadLocation("America/Chicago")
	end, _ := time.ParseInLocation("2006-01-02", params["end_date"][0], chi)
	end = end.AddDate(0, 0, 1) // inc to the following day
	start := end.AddDate(0, 0, -days)

	rows, err := api.Db.Query(`SELECT total,ward,requested_date 
		FROM daily_counts
		WHERE service_code = $1 
			AND requested_date >= $2 
			AND requested_date < $3
		ORDER BY ward DESC, requested_date DESC`,
		string(service_code), start, end)

	if err != nil {
		return backend_error(err)
	}

	data := make(map[int]map[string]int)
	// { 32: { '2013-07-23': 42, '2013-07-24': 41 }, 3: { '2013-07-23': 42, '2013-07-24': 41 } }

	for rows.Next() {
		var ward, count int
		var date time.Time

		if err := rows.Scan(&count, &ward, &date); err != nil {
			return backend_error(err)
		}

		if _, present := data[ward]; !present {
			data[ward] = make(map[string]int)
		}

		data[ward][date.Format("2006-01-02")] = count
	}

	type WardCount struct {
		Ward    int     `json:"ward"`
		Counts  []int   `json:"counts"`
		Average float32 `json:"average"`
	}

	counts := make(map[int]WardCount)
	var day_data []string

	// generate a list of days returned in the results
	for day := 0; day < days; day++ {
		day_data = append(day_data, start.AddDate(0, 0, day).Format("2006-01-02"))
	}

	// for each ward, and each day, find the count and populate result
	for i := 1; i < 51; i++ {
		for day := 0; day < days; day++ {
			d := start.AddDate(0, 0, day)
			c := 0
			if total_for_day, present := data[i][d.Format("2006-01-02")]; present {
				c = total_for_day
			}

			tmp := counts[i]
			tmp.Counts = append(counts[i].Counts, c)
			counts[i] = tmp
		}
	}

	rows, err = api.Db.Query(`SELECT SUM(total)/365.0, ward
             FROM daily_counts
             WHERE requested_date >= DATE(NOW() - INTERVAL '1 year')
                     AND service_code = $1
             GROUP BY ward;`, service_code)

	if err != nil {
		return backend_error(err)
	}

	for rows.Next() {
		var count float32
		var ward int
		if err := rows.Scan(&count, &ward); err != nil {
			return backend_error(err)
		}

		tmp := counts[ward]
		tmp.Average = count
		counts[ward] = tmp
	}

	type CityCount struct {
		Average  float32 `json:"average"`
		DailyMax []int   `json:"daily_max"`
		Count    int     `json:"count"`
	}

	// find total opened for the entire city for date range
	var city_total CityCount
	err = api.Db.QueryRow(`SELECT SUM(total)
                     FROM daily_counts
                     WHERE service_code = $1
                             AND requested_date >= $2
                             AND requested_date < $3;`,
		string(service_code), start, end).Scan(&city_total.Count)

	if err != nil {
		return backend_error(err)
	}

	city_total.Average = float32(city_total.Count) / 365.0

	// find the seven largest days of all time
	rows, err = api.Db.Query(`SELECT total
                     FROM daily_counts
                     WHERE service_code = $1
                     ORDER BY total DESC
                     LIMIT 7;`,
		string(service_code))

	for rows.Next() {
		var daily_max int
		if err := rows.Scan(&daily_max); err != nil {
			return backend_error(err)
		}

		city_total.DailyMax = append(city_total.DailyMax, daily_max)
	}

	// pluck data to return, ensure we return a number, even zero, for each ward
	type WC struct {
		Counts  []int   `json:"counts"`
		Average float32 `json:"average"`
	}

	complete_wards := make(map[string]WC)
	for i := 1; i < 51; i++ {
		k := strconv.Itoa(i)
		tmp := complete_wards[k]
		tmp.Counts = counts[i].Counts
		tmp.Average = counts[i].Average
		complete_wards[k] = tmp
	}

	type RespData struct {
		DayData  []string      `json:"day_data"`
		CityData CityCount     `json:"city_data"`
		WardData map[string]WC `json:"ward_data"`
	}

	return dumpJson(RespData{CityData: city_total, WardData: complete_wards, DayData: day_data}), nil
}
