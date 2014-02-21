package main

import (
  "github.com/gorilla/mux"
  "net/http"
  "net/url"
  "strconv"
  "time"
)

func WardServicesHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
  // for a given ward, return the number of service requests opened and closed,
  // grouped by service_code
  //
  // Parameters:
  //
  //  count:          the number of days of data to return
  //  end_date:       date that +count+ is based from.
  //  callback:       function to wrap response in (for JSONP functionality)
  //
  // Sample API output
  //
  // $ curl "http://localhost:5000/wards/10/services.json?count=7&end_date=2013-08-30"
  // [
  //   {
  //     opened: 5,
  //     closed: 7,
  //     service_code: "4ffa9cad6018277d4000007b"
  //   },
  //   {
  //     opened: 1,
  //     closed: 2,
  //     service_code: "4ffa4c69601827691b000018"
  //   },
  //   {
  //     opened: 33,
  //     closed: 49,
  //     service_code: "4fd3b167e750846744000005"
  //   },
  //   {
  //     opened: 14,
  //     closed: 23,
  //     service_code: "4fd3b656e750846c53000004"
  //   },
  //   {
  //     opened: 7,
  //     closed: 4,
  //     service_code: "4ffa9db16018277d400000a2"
  //   },
  //   {
  //     opened: 14,
  //     closed: 18,
  //     service_code: "4fd3bd3de750846c530000b9"
  //   },
  //   {
  //     opened: 1,
  //     closed: 1,
  //     service_code: "4ffa971e6018277d4000000b"
  //   },
  //   {
  //     opened: 6,
  //     closed: 50,
  //     service_code: "4fd3bbf8e750846c53000069"
  //   },
  //   {
  //     opened: 4,
  //     closed: 5,
  //     service_code: "4fd3b750e750846c5300001d"
  //   },
  //   {
  //     opened: 7,
  //     closed: 5,
  //     service_code: "4ffa9f2d6018277d400000c8"
  //   }
  // ]

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