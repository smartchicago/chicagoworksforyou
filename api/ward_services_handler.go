package main

import (
  "github.com/gorilla/mux"
  "net/http"
  "net/url"
  "strconv"
  "time"
)

func WardServicesHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
  // for a given ward, return the number of service requests opened and closed
  // grouped by day, then by service request type
  //
  // Parameters:
  //
  //  count:          the number of days of data to return
  //  end_date:       date that +count+ is based from.
  //  service_code:   (optional) the code used by the City of Chicago to categorize service requests
  //  callback:       function to wrap response in (for JSONP functionality)
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

  // end = end.AddDate(0, 0, 1) // inc to the following day
  start := end.AddDate(0, 0, -days)

  type ServicesCount struct {
    Opened        int    `json:"opened"`
    Closed        int    `json:"closed"`
    Service_code string `json:"service_code"`
  }

  var services []ServicesCount

  rows, err := api.Db.Query(`SELECT SUM(dc.total) AS opened, SUM(dcc.total) AS closed, service_code
        FROM daily_counts dc
        INNER JOIN daily_closed_counts dcc
        USING(requested_date, ward, service_code)
        WHERE ward = $1
          AND requested_date >= $2
          AND requested_date <= $3
        GROUP BY service_code;`, ward_id, start, end)

  if err != nil {
    return backend_error(err)
  }

  for rows.Next() {
    var opened_count int
    var closed_count int
    var service_code string

    if err := rows.Scan(&opened_count, &closed_count, &service_code); err != nil {
      return backend_error(err)
    }

    row := ServicesCount{Opened: opened_count, Closed: closed_count, Service_code: service_code}
    services = append(services, row)
  }

  return dumpJson(services), nil
}