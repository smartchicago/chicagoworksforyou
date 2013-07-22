package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kylelemons/go-gypsy/yaml"
	"github.com/lib/pq"
	"log"
	"net/http"
	"net/url"
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

var (
	api         Api
	environment = flag.String("environment", "", "Environment to run in, e.g. staging, production")
	config      = flag.String("config", "./config/database.yml", "database configuration file")
	port        = flag.Int("port", 5000, "port that server will listen to (default: 5000)")
)

func init() {
	log.Print("starting ChicagoWorksforYou.com API server")

	// version
	api.Version = "0.9.0"

	// load db config
	flag.Parse()
	log.Printf("running in %s environment, configuration file %s", *environment, *config)
	settings := yaml.ConfigFile(*config)

	// setup database connection
	driver, err := settings.Get(fmt.Sprintf("%s.driver", *environment))
	if err != nil {
		log.Fatal("error loading db driver", err)
	}

	connstr, err := settings.Get(fmt.Sprintf("%s.connstr", *environment))
	if err != nil {
		log.Fatal("error loading db connstr", err)
	}

	db, err := sql.Open(driver, connstr)
	if err != nil {
		log.Fatal("Cannot open database connection", err)
	}

	log.Printf("database connstr: %s", connstr)

	api.Db = db
}

func main() {
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
	router.HandleFunc("/health_check", endpoint(HealthCheckHandler))
	// router.HandleFunc("/services.json", endpoint(ServicesHandler))
	// router.HandleFunc("/requests/time_to_close.json", endpoint(TimeToCloseHandler))
	// router.HandleFunc("/wards/{id}/requests.json", endpoint(WardRequestsHandler))
	// router.HandleFunc("/wards/{id}/counts.json", endpoint(WardCountsHandler))
	// router.HandleFunc("/requests/{service_code}/counts.json", endpoint(RequestCountsHandler))
	// router.HandleFunc("/requests/counts_by_day.json", endpoint(DayCountsHandler))
	// router.HandleFunc("/requests/media.json", endpoint(RequestsMediaHandler))

	log.Printf("CWFY ready for battle on port %d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), router))
}

type ApiEndpoint func(url.Values) ([]byte, *ApiError)
type ApiError struct {
	Msg  string // human readable error message
	Code int    // http status code to use
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("api error %d: %s", e.Code, e.Msg)
}

func endpoint(f ApiEndpoint) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w = setHeaders(w)
		params := req.URL.Query()

		log.Printf("[cwfy %s] %s%s\t%+v", api.Version, req.URL.Host, req.URL.RequestURI(), params)

		t := time.Now()
		response, err := f(params)

		if err != nil {
			log.Printf(err.Error())
			http.Error(w, err.Msg, err.Code)			
		}

		w.Write(response)
		diff := time.Now()
		log.Printf("[cwfy %s] %s%s completed in %v", api.Version, req.URL.Host, req.URL.RequestURI(), diff.Sub(t))
	}
}

func setHeaders(w http.ResponseWriter) http.ResponseWriter {
	// set HTTP headers on the response object
	// TODO: add cache control headers

	w.Header().Set("Content-type", "application/json; charset=utf-8")
	w.Header().Set("Server", fmt.Sprintf("ChicagoWorksForYou.com/%s", api.Version))
	return w
}

func dumpJson(in interface{}) []byte {
	out, err := json.MarshalIndent(in, "", "  ")
	if err != nil {
		log.Printf("error marshalling to json: %s", err)
	}
	return out
}

func WrapJson(unwrapped []byte, callback []string) (jsn []byte) {
	jsn = unwrapped
	if len(callback) > 0 {
		wrapped := strings.Join([]string{callback[0], "(", string(jsn), ");"}, "")
		jsn = []byte(wrapped)
	}

	return
}

// func HealthCheckHandler(response http.ResponseWriter, request *http.Request) {
func HealthCheckHandler(params url.Values) ([]byte, *ApiError) {
	// params := request.URL.Query()

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

	return dumpJson(health_check), nil
}

func RequestsMediaHandler(response http.ResponseWriter, request *http.Request) {
	// Return 500 most recent SR that have media "attached"
	//
	// Sample:
	//
	// $ curl "http://localhost:5000/requests/media.json"
	// [
	//   {
	//     "Service_name": "Graffiti Removal",
	//     "Address": "1000 W Cullerton St Pilsen",
	//     "Media_url": "http://311request.cityofchicago.org/media/chicago/report/photos/51586114016382d1fed662a3/image.jpg",
	//     "Service_request_id": "13-00358810",
	//     "Ward": 25
	//   },
	//   {
	//     "Service_name": "Sanitation Code Violation",
	//     "Address": "2168 n parkside ave",
	//     "Media_url": "http://311request.cityofchicago.org/media/chicago/report/photos/516086be0163865707dd2e40/pic_5023_2111.png",
	//     "Service_request_id": "13-00389663",
	//     "Ward": 29
	//   },
	//   {
	//     "Service_name": "Graffiti Removal",
	//     "Address": "2133 S Union Ave East Pilsen",
	//     "Media_url": "http://311request.cityofchicago.org/media/chicago/report/photos/5161b7cb0163865707dd48f2/area.png",
	//     "Service_request_id": "13-00391264",
	//     "Ward": 25
	//   },

	params := request.URL.Query()

	type SR struct {
		Service_name, Address, Media_url, Service_request_id string
		Ward                                                 int
	}

	var sr_with_media []SR

	rows, err := api.Db.Query(`SELECT service_name,address,media_url,service_request_id,ward 
                FROM service_requests
                WHERE media_url != ''
                ORDER BY requested_datetime DESC
                LIMIT 500;`)

	if err != nil {
		log.Print("error laoding media objects", err)
	}

	for rows.Next() {
		sr := SR{}
		if err := rows.Scan(&sr.Service_name, &sr.Address, &sr.Media_url, &sr.Service_request_id, &sr.Ward); err != nil {
			log.Print("error", err)
		}

		sr_with_media = append(sr_with_media, sr)
	}

	jsn, _ := json.MarshalIndent(sr_with_media, "", "  ")
	jsn = WrapJson(jsn, params["callback"])

	response.Write(jsn)
}

func DayCountsHandler(response http.ResponseWriter, request *http.Request) {
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

	params := request.URL.Query()

	chi, _ := time.LoadLocation("America/Chicago")
	end, _ := time.ParseInLocation("2006-01-02", params["day"][0], chi)
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
		log.Print("error loading day counts: ", err)
	}

	for rows.Next() {
		var dc DayCount
		var sc string
		if err := rows.Scan(&sc, &dc.Count); err != nil {
			log.Print("error loading daily counts from DB", err)
		}
		counts[sc] = dc
	}

	// fetch top ward(s) for each service_code
	// for each service code, find the ward (or wards) with the most SR opened

	// for each service code, fetch wards for day sorted by # SR opened

	service_codes := []string{"4fd3bd72e750846c530000cd", "4ffa9cad6018277d4000007b", "4ffa4c69601827691b000018", "4fd3b167e750846744000005", "4fd3b656e750846c53000004", "4ffa971e6018277d4000000b", "4fd3bd3de750846c530000b9", "4fd6e4ece750840569000019", "4fd3b9bce750846c5300004a", "4ffa9db16018277d400000a2", "4ffa995a6018277d4000003c", "4fd3bbf8e750846c53000069", "4fd3b750e750846c5300001d", "4ffa9f2d6018277d400000c8"} //FIXME: don't hard code this

	for _, sc := range service_codes {
		wards := make(map[int]int) // map the ward to its total number of reqs for the service code for the day

		rows, err := api.Db.Query(`SELECT total, ward 
                     FROM daily_counts
                     WHERE requested_date >= $1 
                             AND requested_date < $2
                             AND service_code = $3
                     ORDER BY total DESC;`, start, end, sc)

		if err != nil {
			log.Print("error loading top wards: ", err)
		}

		for rows.Next() {
			var ward, total int
			if err := rows.Scan(&total, &ward); err != nil {
				log.Print("error loading daily counts from DB", err)
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
	rows, err = api.Db.Query(`SELECT service_code, AVG(total) AS avg_reports 
		FROM daily_counts 
		WHERE requested_date >= (NOW() - INTERVAL '1 year')
		GROUP BY service_code
		ORDER BY avg_reports;`)

	if err != nil {
		log.Print("error loading averages", err)
	}

	for rows.Next() {
		var sc string
		var avg float32
		if err := rows.Scan(&sc, &avg); err != nil {
			// handle
		}

		tmp := counts[sc]
		tmp.Average = avg
		counts[sc] = tmp
	}

	jsn, _ := json.MarshalIndent(counts, "", "  ")
	jsn = WrapJson(jsn, params["callback"])

	response.Write(jsn)
}

func RequestCountsHandler(response http.ResponseWriter, request *http.Request) {
	// for a given request service type and date, return the count
	// of requests for that date, grouped by ward, and the city total
	// The output is a map where keys are ward identifiers, and the value is the count.
	// The city total for the time interval is assigned to ward #0
	//
	// Sample request and output:
	// $ curl "http://localhost:5000/requests/4fd3b167e750846744000005/counts.json?end_date=2013-06-10&count=1"
	// {
	// 	  "0": {
	// 	    "Count": 1107,
	// 	    "Average": 3.0328767
	// 	  },
	// 	  "1": {
	// 	    "Count": 63,
	// 	    "Average": 17.495846
	// 	  },
	// 	  "10": {
	// 	    "Count": 21,
	// 	    "Average": 7.6055045
	// 	  },

	vars := mux.Vars(request)
	service_code := vars["service_code"]
	params := request.URL.Query()

	// determine date range. default is last 7 days.
	days, _ := strconv.Atoi(params["count"][0])

	chi, _ := time.LoadLocation("America/Chicago")
	end, _ := time.ParseInLocation("2006-01-02", params["end_date"][0], chi)
	end = end.AddDate(0, 0, 1) // inc to the following day
	start := end.AddDate(0, 0, -days)

	rows, err := api.Db.Query(`SELECT SUM(total), ward 
		FROM daily_counts
		WHERE service_code = $1 
			AND requested_date >= $2 
			AND requested_date < $3
		GROUP BY ward 
		ORDER BY ward`,
		string(service_code), start, end)

	if err != nil {
		log.Fatal("error fetching data for RequestCountsHandler", err)
	}

	type WardCount struct {
		Ward    int
		Count   int
		Average float32
	}

	counts := make(map[int]WardCount)
	for rows.Next() {
		wc := WardCount{}
		if err := rows.Scan(&wc.Count, &wc.Ward); err != nil {
			log.Print("error reading row of ward count", err)
		}

		// trunc the requested time to just date
		counts[wc.Ward] = wc
	}

	rows, err = api.Db.Query(`SELECT AVG(total), ward 
		FROM daily_counts 
		WHERE requested_date >= DATE(NOW() - INTERVAL '1 year') 
			AND service_code = $1 
		GROUP BY ward;`, service_code)

	if err != nil {
		log.Print("error querying for year counts", err)
	}

	for rows.Next() {
		var count float32
		var ward int
		if err := rows.Scan(&count, &ward); err != nil {
			log.Print("error loading ward counts ", err, count, ward)
		}

		tmp := counts[ward]
		tmp.Average = count
		counts[ward] = tmp
	}

	// find total opened for the entire city for date range
	city_total := WardCount{Ward: 0, Count: 0, Average: 0.0}
	err = api.Db.QueryRow(`SELECT SUM(total) 
		FROM daily_counts 
		WHERE service_code = $1 
			AND requested_date >= $2 
			AND requested_date < $3;`,
		string(service_code), start, end).Scan(&city_total.Count)

	if err != nil {
		log.Print("error loading city-wide total count for %s. err: %s", service_code, err)
	}

	city_total.Average = float32(city_total.Count) / 365.0
	counts[0] = city_total

	log.Printf("city total: %+v", city_total)

	// pluck data to return, ensure we return a number, even zero, for each ward
	type WC struct {
		Count   int
		Average float32
	}

	data := make(map[string]WC)
	for i := 0; i < 51; i++ {
		k := strconv.Itoa(i)
		tmp := data[k]
		tmp.Count = counts[i].Count
		tmp.Average = counts[i].Average
		data[k] = tmp
	}

	jsn, _ := json.MarshalIndent(data, "", "  ")
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

	chi, _ := time.LoadLocation("America/Chicago")
	end, _ := time.ParseInLocation("2006-01-02", params["end_date"][0], chi)
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

	// zero init the times map
	for i := 1; i < 51; i++ {
		times[strconv.Itoa(i)] = TimeToClose{Time: 0.0, Total: 0, Ward: i}
	}

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
	params := request.URL.Query()

	// determine date range.
	days, _ := strconv.Atoi(params["count"][0])

	chi, _ := time.LoadLocation("America/Chicago")
	end, _ := time.ParseInLocation("2006-01-02", params["end_date"][0], chi)
	end = end.AddDate(0, 0, 1) // inc to the following day
	start := end.AddDate(0, 0, -days)

	service_code := params["service_code"][0]

	rows, err := api.Db.Query(`SELECT COUNT(*), DATE(requested_datetime) AS requested_date 
		FROM service_requests 
		WHERE ward = $1
			AND duplicate IS NULL 
			AND service_code = $2 
			AND requested_datetime >= $3::date 
			AND requested_datetime <= $4::date
		GROUP BY DATE(requested_datetime) 
		ORDER BY requested_date;`,
		string(ward_id), service_code, start, end)

	if err != nil {
		log.Fatal("error fetching data for WardCountsHandler", err)
	}

	type WardCount struct {
		Count       int
		CityTotal   int
		CityAverage float32
	}

	counts := make(map[string]WardCount)
	for rows.Next() {
		var wc WardCount
		var rd time.Time

		if err := rows.Scan(&wc.Count, &rd); err != nil {
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

	jsn, _ := json.MarshalIndent(resp, "", "  ")
	jsn = WrapJson(jsn, params["callback"])

	response.Write(jsn)
}

func WardRequestsHandler(response http.ResponseWriter, request *http.Request) {
	// for a given ward, return recent service requests
	vars := mux.Vars(request)
	ward_id := vars["id"]
	params := request.URL.Query()

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
