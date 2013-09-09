package main

import (
	"database/sql"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func TimeToCloseHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
	// Given service type, date, length of time & increment,
	// return time-to-close for that service type, for each
	// increment over that length of time, going backwards from that date.
	// If service_code is omitted, the average will be for all service types.
	//
	// Response data:
	//      The city-wide average will be returned in the CityData map.
	//      "Count" is the number of service requests closed in the given time period.
	//      "Time" is the average difference, in days, between closed and requested datetimes.
	//
	// Sample request and output:
	// $ curl "http://localhost:5000/requests/time_to_close.json?end_date=2013-06-19&count=7&service_code=4fd3b167e750846744000005"
	// {
	//   "WardData": {
	//     "1": {
	//       "Time": 6.586492353553241,
	//       "Count": 643
	//     },
	// 	( .. truncated ...)
	//     "9": {
	//       "Time": 2.469373385011574,
	//       "Count": 43
	//     }
	//   },
	//   "CityData": {
	//     "Time": 3.8197868124884256,
	//     "Count": 11123
	//   },
	//   "Threshold": 27.537741650677532
	// }

	// required
	service_code := params.Get("service_code")
	days, err := strconv.Atoi(params.Get("count"))
	if err != nil || days > 60 || days < 1 {
		return nil, &ApiError{Msg: "invalid count, must be integer, 1..60", Code: 400}
	}

	chi, _ := time.LoadLocation("America/Chicago")
	end, err := time.ParseInLocation("2006-01-02", params.Get("end_date"), chi)
	if err != nil {
		return nil, &ApiError{Msg: "invalid end_date", Code: 400}
	}

	end = end.AddDate(0, 0, 1) // inc to the following day
	start := end.AddDate(0, 0, -days)

	var rows *sql.Rows

	if service_code != "" {
		rows, err = api.Db.Query(`SELECT EXTRACT('EPOCH' FROM AVG(closed_datetime - requested_datetime)) AS avg_ttc, COUNT(service_request_id), ward
        		FROM service_requests 
        		WHERE closed_datetime IS NOT NULL 
        			AND duplicate IS NULL
        			AND closed_datetime >= $1
        			AND closed_datetime <= $2
        			AND ward IS NOT NULL
        			AND service_code = $3 
        		GROUP BY ward 
        		ORDER BY avg_ttc DESC;`, start, end, service_code)
	} else {
		rows, err = api.Db.Query(`SELECT EXTRACT('EPOCH' FROM AVG(closed_datetime - requested_datetime)) AS avg_ttc, COUNT(service_request_id), ward
        		FROM service_requests 
        		WHERE closed_datetime IS NOT NULL 
        			AND duplicate IS NULL
        			AND closed_datetime >= $1
        			AND closed_datetime <= $2
        			AND ward IS NOT NULL
        		GROUP BY ward 
        		ORDER BY avg_ttc DESC;`, start, end)
	}

	if err != nil {
		return backend_error(err)
	}

	type TimeToClose struct {
		Time  float64
		Count int
		Ward  int `json:"-"`
	}

	times := make(map[string]TimeToClose)

	// zero init the times map
	for i := 1; i < 51; i++ {
		times[strconv.Itoa(i)] = TimeToClose{Time: 0.0, Count: 0, Ward: i}
	}

	for rows.Next() {
		var ttc TimeToClose
		if err := rows.Scan(&ttc.Time, &ttc.Count, &ttc.Ward); err != nil {
			return backend_error(err)
		}
		ttc.Time = ttc.Time / 86400.0 // convert from seconds to days
		times[strconv.Itoa(ttc.Ward)] = ttc
	}

	// find the city-wide average for the interval/service code
	city_average := TimeToClose{Ward: 0}

	if service_code != "" {
		err = api.Db.QueryRow(`SELECT EXTRACT('EPOCH' FROM AVG(closed_datetime - requested_datetime)) AS avg_ttc, COUNT(service_request_id)
        		FROM service_requests 
        		WHERE closed_datetime IS NOT NULL 
        			AND duplicate IS NULL
        			AND service_code = $1 
        			AND closed_datetime >= $2
        			AND closed_datetime <= $3
        			AND ward IS NOT NULL`, service_code, start, end).Scan(&city_average.Time, &city_average.Count)

	} else {
		err = api.Db.QueryRow(`SELECT EXTRACT('EPOCH' FROM AVG(closed_datetime - requested_datetime)) AS avg_ttc, COUNT(service_request_id)
        		FROM service_requests 
        		WHERE closed_datetime IS NOT NULL 
        			AND duplicate IS NULL
        			AND closed_datetime >= $1
        			AND closed_datetime <= $2
        			AND ward IS NOT NULL`, start, end).Scan(&city_average.Time, &city_average.Count)
	}

	if err != nil {
		return backend_error(err)
	}

	city_average.Time = city_average.Time / 86400.0 // convert to days

	// calculate bottom threshold of values to display
	var std_dev, sum float64
	for i := 1; i < 51; i++ {
		sum += math.Pow((float64(times[strconv.Itoa(i)].Count) - (float64(city_average.Count) / 50.0)), 2)
	}

	std_dev = math.Sqrt(sum / 50.0)
	threshold := (float64(city_average.Count) / 50.0) - std_dev

	type resp_data struct {
		WardData  map[string]TimeToClose
		CityData  TimeToClose
		Threshold float64
	}

	return dumpJson(resp_data{WardData: times, CityData: city_average, Threshold: threshold}), nil
}
