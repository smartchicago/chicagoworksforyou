package main

import (
	"database/sql"
	"encoding/json"
	"github.com/bmizerany/pq"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

type Api struct {
	Db      *sql.DB
	Version string
}

var api Api

func init() {
	// version
	api.Version = "0.0.1"

	// setup database connection
	db, err := sql.Open("postgres", "dbname=cwfy sslmode=disable")
	if err != nil {
		log.Fatal("Cannot open database connection", err)
	}
	api.Db = db
}

func main() {
	log.Print("starting ChicagoWorksforYou.com API server")

	// listen for SIGINT (h/t http://stackoverflow.com/a/12571099/1247272)
	notify_channel := make(chan os.Signal, 1)
	signal.Notify(notify_channel, os.Interrupt, os.Kill)
	go func() {
		for _ = range notify_channel {
			log.Printf("stopping ChicagoWorksForYou.com API server")
			api.Db.Close()
			os.Exit(1)
		}
	}()

	router := mux.NewRouter()
	router.HandleFunc("/health_check", HealthCheckHandler)
	router.HandleFunc("/services.json", ServicesHandler)
	router.HandleFunc("/requests/time_to_close.json", TimeToCloseHandler)
	router.HandleFunc("/wards/{id}/requests.json", WardRequestsHandler)
	router.HandleFunc("/wards/{id}/counts.json", WardCountsHandler)
	router.HandleFunc("/requests/{service_code}/counts.json", RequestCountsHandler)
	http.ListenAndServe(":5000", router)
}

func WrapJson(unwrapped []byte, callback []string) (jsn []byte) {
	jsn = unwrapped
	if len(callback) > 0 {
		wrapped := strings.Join([]string{callback[0], "(", string(jsn), ");"}, "")
		jsn = []byte(wrapped)
	}

	return
}

func RequestCountsHandler(response http.ResponseWriter, request *http.Request) {
	// for a given request service type and date, return the count
	// of requests for that date, grouped by ward

	vars := mux.Vars(request)
	service_code := vars["service_code"]
	params := request.URL.Query()

	// determine date range. default is last 7 days.
	days, _ := strconv.Atoi(params["count"][0])

	end, _ := time.Parse("2006-01-02", params["end_date"][0])
	end = end.AddDate(0, 0, 1) // inc to the following day
	start := end.AddDate(0, 0, -days)

	log.Printf("date range is %s to %s", start, end)

	rows, err := api.Db.Query("SELECT COUNT(*), ward FROM service_requests WHERE service_code "+
		"= $1 AND duplicate IS NULL AND requested_datetime >= $2::date "+
		" AND requested_datetime <= $3::date  GROUP BY ward ORDER BY ward;",
		string(service_code), start, end)

	if err != nil {
		log.Fatal("error fetching data for RequestCountsHandler", err)
	}

	type WardCount struct {
		Ward  int
		Count int
	}

	counts := make(map[string]WardCount)
	for rows.Next() {
		wc := WardCount{}
		if err := rows.Scan(&wc.Count, &wc.Ward); err != nil {
			log.Print("error reading row of ward count", err)
		}

		// trunc the requested time to just date
		counts[strconv.Itoa(wc.Ward)] = wc
	}

	// find total opened for the entire city for date range
	city_total := WardCount{Ward: 0}
	err = api.Db.QueryRow("SELECT COUNT(*) FROM service_requests WHERE service_code "+
		"= $1 AND duplicate IS NULL AND requested_datetime >= $2::date "+
		" AND requested_datetime <= $3::date;",
		string(service_code), start, end).Scan(&city_total.Count)

	if err != nil {
		log.Print("error loading city-wide total count for %s. err: %s", service_code, err)
	}

	counts["0"] = city_total

	jsn, _ := json.MarshalIndent(counts, "", "  ")
	jsn = WrapJson(jsn, params["callback"])

	response.Write(jsn)
}

func TimeToCloseHandler(response http.ResponseWriter, request *http.Request) {
	// Given service type, date, length of time & increment,
	// return time-to-close for that service type, for each
	// increment over that length of time, going backwards from that date.
	//
	// Response data:
	//      The city-wide average will be returned as ward "0".
	//      "Total" is the number of service requests closed in the given time period.
	//      "Time" is the average difference, in days, between closed and requested datetimes.
	//      NOTE: This value may be negative, due to wonky data from the City. Go figure.
	//
	// Sample request and output:
	// $ curl "http://localhost:5000/requests/time_to_close.json?end_date=2013-06-19&count=7&service_code=4fd3b167e750846744000005"
	//        {
	//          "0": {
	//            "Time": 0.8193135,
	//            "Total": 275,
	//            "Ward": 0
	//          },
	//          "1": {
	//            "Time": 1.0570672,
	//            "Total": 5,
	//            "Ward": 1
	//          },
	//          "11": {
	//            "Time": -0.015823688,
	//            "Total": 12,
	//            "Ward": 11
	//          },
	//          "12": {
	//            "Time": -0.0120927375,
	//            "Total": 16,
	//            "Ward": 12
	//          },
	//      ... snipped ...

	params := request.URL.Query()

	// required
	service_code := params["service_code"][0]
	days, _ := strconv.Atoi(params["count"][0])

	end, _ := time.Parse("2006-01-02", params["end_date"][0])
	end = end.AddDate(0, 0, 1) // inc to the following day
	start := end.AddDate(0, 0, -days)

	rows, err := api.Db.Query("SELECT EXTRACT('EPOCH' FROM AVG(closed_datetime - requested_datetime)) AS avg_ttc, COUNT(service_request_id), ward "+
		"FROM service_requests WHERE closed_datetime IS NOT NULL AND duplicate IS NULL "+
		"AND service_code = $1 AND closed_datetime >= $2 AND closed_datetime <= $3"+
		"GROUP BY ward ORDER BY avg_ttc DESC;", service_code, start, end)

	if err != nil {
		log.Print("error fetching time to close", err)
	}

	type TimeToClose struct {
		Time  float32
		Total int
		Ward  int
	}

	times := make(map[string]TimeToClose)

	for rows.Next() {
		var ttc TimeToClose
		if err := rows.Scan(&ttc.Time, &ttc.Total, &ttc.Ward); err != nil {
			log.Print("error loading time to close counts", err)
		}
		ttc.Time = ttc.Time / 86400.0 // convert from seconds to days
		times[strconv.Itoa(ttc.Ward)] = ttc
	}

	// find the city-wide average for the interval/service code
	city_average := TimeToClose{Ward: 0}
	err = api.Db.QueryRow("SELECT EXTRACT('EPOCH' FROM AVG(closed_datetime - requested_datetime)) AS avg_ttc, COUNT(service_request_id)"+
		"FROM service_requests WHERE closed_datetime IS NOT NULL AND duplicate IS NULL "+
		"AND service_code = $1 AND closed_datetime >= $2 AND closed_datetime <= $3",
		service_code, start, end).Scan(&city_average.Time, &city_average.Total)

	if err != nil {
		log.Print("error fetching city average time to close", err)
	}

	city_average.Time = city_average.Time / 86400.0 // convert to days
	times["0"] = city_average

	jsn, err := json.MarshalIndent(times, "", "  ")
	if err != nil {
		log.Print("error marshaling to JSON", err)
	}
	jsn = WrapJson(jsn, params["callback"])

	response.Write(jsn)
}

func WardCountsHandler(response http.ResponseWriter, request *http.Request) {
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
	// Note that the end date is June 12, and the results include the end_date.
	// $ curl "http://localhost:5000/wards/10/counts.json?service_code=4fd3b167e750846744000005&count=5&end_date=2013-06-12"
	// {
	//   "2013-06-06": 2,
	//   "2013-06-07": 4,
	//   "2013-06-09": 5,
	//   "2013-06-10": 6,
	//   "2013-06-12": 23
	// }
	//
	vars := mux.Vars(request)
	ward_id := vars["id"]
	params := request.URL.Query()

	// determine date range.
	days, _ := strconv.Atoi(params["count"][0])

	end, _ := time.Parse("2006-01-02", params["end_date"][0])
	end = end.AddDate(0, 0, 1) // inc to the following day
	start := end.AddDate(0, 0, -days)

	log.Printf("fetching counts for ward %s code %s for past %d days", ward_id, params["service_code"][0], days)
	log.Printf("date range is %s to %s", start, end)

	rows, err := api.Db.Query("SELECT COUNT(*), DATE(requested_datetime) as requested_date FROM service_requests WHERE ward = $1 "+
		"AND duplicate IS NULL AND service_code = $2 AND requested_datetime >= $3::date AND requested_datetime <= $4::date "+
		"GROUP BY DATE(requested_datetime) ORDER BY requested_date;", string(ward_id), params["service_code"][0], start, end)
	if err != nil {
		log.Fatal("error fetching data for WardCountsHandler", err)
	}

	type WardCount struct {
		Requested_date time.Time
		Count          int
	}

	var counts []WardCount
	for rows.Next() {
		wc := WardCount{}
		if err := rows.Scan(&wc.Count, &wc.Requested_date); err != nil {
			log.Print("error reading row of ward count", err)
		}

		// trunc the requested time to just date
		counts = append(counts, wc)
	}

	resp := make(map[string]int)

	for _, c := range counts {
		key := c.Requested_date.Format("2006-01-02")
		resp[key] = c.Count
	}

	jsn, _ := json.MarshalIndent(resp, "", "  ")
	jsn = WrapJson(jsn, params["callback"])

	response.Write(jsn)
}

func WardRequestsHandler(response http.ResponseWriter, request *http.Request) {
	// for a given ward, return recent service requests

	vars := mux.Vars(request)
	ward_id := vars["id"]
	params := request.URL.Query()
	log.Print("fetch requests for ward ", ward_id)

	rows, err := api.Db.Query("SELECT lat,long,ward,police_district,service_request_id,status,service_name,service_code,agency_responsible,address,channel,media_url,requested_datetime,updated_datetime,created_at,updated_at,duplicate,parent_service_request_id,id FROM service_requests WHERE duplicate IS NULL AND ward = $1 ORDER BY updated_at DESC LIMIT 100;", ward_id)

	if err != nil {
		log.Fatal("error fetching data for WardRequestsHandler", err)
	}

	type Open311RequestRow struct {
		Lat, Long                                                                                                                                     float64
		Ward, Police_district, Id                                                                                                                     int
		Service_request_id, Status, Service_name, Service_code, Agency_responsible, Address, Channel, Media_url, Duplicate, Parent_service_request_id sql.NullString
		Requested_datetime, Updated_datetime, Created_at, Updated_at                                                                                  pq.NullTime // FIXME: should these be proper time objects?
		Extended_attributes                                                                                                                           map[string]interface{}
	}

	var result []Open311RequestRow

	for rows.Next() {
		var row Open311RequestRow
		if err := rows.Scan(&row.Lat, &row.Long, &row.Ward, &row.Police_district,
			&row.Service_request_id, &row.Status, &row.Service_name,
			&row.Service_code, &row.Agency_responsible, &row.Address,
			&row.Channel, &row.Media_url, &row.Requested_datetime,
			&row.Updated_datetime, &row.Created_at, &row.Updated_at,
			&row.Duplicate, &row.Parent_service_request_id,
			&row.Id); err != nil {
			log.Fatal("error reading row", err)
		}

		result = append(result, row)
	}

	jsn, _ := json.MarshalIndent(result, "", "  ")
	jsn = WrapJson(jsn, params["callback"])
	response.Write(jsn)
}

func ServicesHandler(response http.ResponseWriter, request *http.Request) {
	// return counts of requests, grouped by service name
	//
	// Sample output:
	//
	// [
	//   {
	//     "Count": 1139,
	//     "Service_code": "4fd3b167e750846744000005",
	//     "Service_name": "Graffiti Removal"
	//   },
	//   {
	//     "Count": 25,
	//     "Service_code": "4fd6e4ece750840569000019",
	//     "Service_name": "Restaurant Complaint"
	//   },
	//
	//  ... snip ...
	//
	// ]

	type ServicesCount struct {
		Count        int
		Service_code string
		Service_name string
	}

	var services []ServicesCount

	rows, err := api.Db.Query("SELECT COUNT(*), service_code, service_name FROM service_requests WHERE duplicate IS NULL GROUP BY service_code,service_name;")

	if err != nil {
		log.Fatal("error fetching data for ServicesHandler", err)
	}

	for rows.Next() {
		var count int
		var service_code, service_name string

		if err := rows.Scan(&count, &service_code, &service_name); err != nil {
			log.Fatal("error reading row", err)
		}

		row := ServicesCount{Count: count, Service_code: service_code, Service_name: service_name}
		services = append(services, row)
	}

	jsn, _ := json.MarshalIndent(services, "", "  ")
	response.Write(jsn)
}

func HealthCheckHandler(response http.ResponseWriter, request *http.Request) {
	params := request.URL.Query()

	response.Header().Add("Content-type", "application/json")

	type HealthCheck struct {
		Count    int
		Database bool
		Healthy  bool
		Version  string
	}

	health_check := HealthCheck{Version: api.Version}

	health_check.Database = api.Db.Ping() == nil

	rows, _ := api.Db.Query("SELECT COUNT(*) FROM service_requests;")
	for rows.Next() {
		if err := rows.Scan(&health_check.Count); err != nil {
			log.Fatal("error fetching count", err)
		}
	}

	// calculate overall health
	health_check.Healthy = health_check.Count > 0 && health_check.Database

	log.Printf("health_check: %+v", health_check)
	if !health_check.Healthy {
		log.Printf("health_check failed")
	}
	jsn, _ := json.MarshalIndent(health_check, "", "  ")
	jsn = WrapJson(jsn, params["callback"])
	response.Write(jsn)
}
