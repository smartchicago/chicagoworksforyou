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
	"math"
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
	router.HandleFunc("/services.json", endpoint(ServicesHandler))
	router.HandleFunc("/requests/time_to_close.json", endpoint(TimeToCloseHandler))
	router.HandleFunc("/wards/{id}/requests.json", endpoint(WardRequestsHandler))
	router.HandleFunc("/wards/{id}/counts.json", endpoint(WardCountsHandler))
	router.HandleFunc("/wards/{id}/historic_highs.json", endpoint(WardHistoricHighsHandler))
	router.HandleFunc("/requests/{service_code}/counts.json", endpoint(RequestCountsHandler))
	router.HandleFunc("/requests/counts_by_day.json", endpoint(DayCountsHandler))
	router.HandleFunc("/requests/media.json", endpoint(RequestsMediaHandler))

	log.Printf("CWFY ready for battle on port %d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), router))
}

type ApiEndpoint func(url.Values, *http.Request) ([]byte, *ApiError)
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

		log.Printf("[cwfy %s] %s %s%s %+v", api.Version, req.RemoteAddr, req.Host, req.RequestURI, params)

		t := time.Now()
		response, err := f(params, req)

		if err != nil {
			log.Printf(err.Error())
			http.Error(w, err.Msg, err.Code)
		}

		w.Write(WrapJson(response, params["callback"]))
		diff := time.Now()
		log.Printf("[cwfy %s] %s %s%s completed in %v", api.Version, req.RemoteAddr, req.Host, req.RequestURI, diff.Sub(t))
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
func HealthCheckHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
	type HealthCheck struct {
		Count             int
		Database, Healthy bool
		Version           string
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

func RequestsMediaHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
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
		log.Print("error laoding media objects ", err)
	}

	for rows.Next() {
		sr := SR{}
		if err := rows.Scan(&sr.Service_name, &sr.Address, &sr.Media_url, &sr.Service_request_id, &sr.Ward); err != nil {
			log.Print("error ", err)
		}

		sr_with_media = append(sr_with_media, sr)
	}

	return dumpJson(sr_with_media), nil
}

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
	rows, err = api.Db.Query(`SELECT service_code, SUM(total)/365.0 AS avg_reports 
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

	return dumpJson(counts), nil
}

func RequestCountsHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
	// for a given request service type and date, return the count
	// of requests for that date, grouped by ward, and the city total
	// The output is a map where keys are ward identifiers, and the value is the count.
	// The city total for the time interval is assigned to ward #0
	//
	// Sample request and output:
	// $ curl "http://localhost:5000/requests/4fd3b167e750846744000005/counts.json?end_date=2013-06-10&count=1"
	// {
	//   "DayData": [
	//     "2013-06-04",
	//     "2013-06-05",
	//     "2013-06-06",
	//     "2013-06-07",
	//     "2013-06-08",
	//     "2013-06-09",
	//     "2013-06-10"
	//   ],
	//   "CityData": {
	//     "Average": 8.084931,
	//     "Count": 2951
	//   },
	//   "WardData": {
	//     "1": {
	//       "Counts": [
	//         29,
	//         19,
	//         40,
	//         60,
	//         16,
	//         2,
	//         35
	//       ],
	//       "Average": 16.671232
	//     },
	//     "10": {
	//       "Counts": [
	//         22,
	//         2,
	//         28,
	//         6,
	//         2,
	//         5,
	//         6
	//       ],
	//       "Average": 6.60274
	//     },

	vars := mux.Vars(request)
	service_code := vars["service_code"]

	// determine date range. default is last 7 days.
	days, _ := strconv.Atoi(params["count"][0])

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
		log.Fatal("error fetching data for RequestCountsHandler", err)
	}

	data := make(map[int]map[string]int)
	// { 32: { '2013-07-23': 42, '2013-07-24': 41 }, 3: { '2013-07-23': 42, '2013-07-24': 41 } }

	for rows.Next() {
		var ward, count int
		var date time.Time

		if err := rows.Scan(&count, &ward, &date); err != nil {
			// FIXME: handle
		}

		if _, present := data[ward]; !present {
			data[ward] = make(map[string]int)
		}

		data[ward][date.Format("2006-01-02")] = count
	}

	// log.Printf("data\n\n%+v", data)

	type WardCount struct {
		Ward    int
		Counts  []int
		Average float32
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

	// log.Printf("counts\n\n%+v", counts)

	rows, err = api.Db.Query(`SELECT SUM(total)/365.0, ward
             FROM daily_counts
             WHERE requested_date >= DATE(NOW() - INTERVAL '1 year')
                     AND service_code = $1
             GROUP BY ward;`, service_code)

	if err != nil {
		log.Print("error querying for year average", err)
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

	type CityCount struct {
		Average float32
		Count   int
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
		log.Print("error loading city-wide total count for %s. err: %s", service_code, err)
	}

	city_total.Average = float32(city_total.Count) / 365.0

	// pluck data to return, ensure we return a number, even zero, for each ward
	type WC struct {
		Counts  []int
		Average float32
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
		DayData  []string
		CityData CityCount
		WardData map[string]WC
	}

	return dumpJson(RespData{CityData: city_total, WardData: complete_wards, DayData: day_data}), nil
}

func TimeToCloseHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
	// Given service type, date, length of time & increment,
	// return time-to-close for that service type, for each
	// increment over that length of time, going backwards from that date.
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
	service_code := params["service_code"][0]
	days, _ := strconv.Atoi(params["count"][0])

	chi, _ := time.LoadLocation("America/Chicago")
	end, _ := time.ParseInLocation("2006-01-02", params["end_date"][0], chi)
	end = end.AddDate(0, 0, 1) // inc to the following day
	start := end.AddDate(0, 0, -days)

	rows, err := api.Db.Query(`SELECT EXTRACT('EPOCH' FROM AVG(closed_datetime - requested_datetime)) AS avg_ttc, COUNT(service_request_id), ward
		FROM service_requests 
		WHERE closed_datetime IS NOT NULL 
			AND duplicate IS NULL
			AND service_code = $1 
			AND closed_datetime >= $2 
			AND closed_datetime <= $3
			AND ward IS NOT NULL
		GROUP BY ward 
		ORDER BY avg_ttc DESC;`, service_code, start, end)

	if err != nil {
		log.Print("error fetching time to close", err)
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
			log.Print("error loading time to close counts", err)
		}
		ttc.Time = ttc.Time / 86400.0 // convert from seconds to days
		times[strconv.Itoa(ttc.Ward)] = ttc
	}

	// find the city-wide average for the interval/service code
	city_average := TimeToClose{Ward: 0}
	err = api.Db.QueryRow(`SELECT EXTRACT('EPOCH' FROM AVG(closed_datetime - requested_datetime)) AS avg_ttc, COUNT(service_request_id)
		FROM service_requests 
		WHERE closed_datetime IS NOT NULL 
			AND duplicate IS NULL
			AND service_code = $1 
			AND closed_datetime >= $2
			AND closed_datetime <= $3
			AND ward IS NOT NULL`, service_code, start, end).Scan(&city_average.Time, &city_average.Count)

	if err != nil {
		log.Print("error fetching city average time to close", err)
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

func WardHistoricHighsHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
	// given a ward and service type, return the set of days with the most SR opened
	//
	// Parameters:
	// 	count: 		number of historicl high days to return.
	//	service_code:   the code used by the City of Chicago to categorize service requests
	//	callback:       function to wrap response in (for JSONP functionality)
	// 	include_today:  if equal to "true" or "1", include the current day (in Chicago) counts as the first element of the result set
	//			Note: if set to true, the number of results returned will be count + 1.
	//

	vars := mux.Vars(request)
	ward_id := vars["id"]

	days, _ := strconv.Atoi(params["count"][0])
	service_code := params["service_code"][0]

	include_today := false
	if val, present := params["include_today"]; present {
		if val[0] == "true" || val[0] == "1" {
			include_today = true
		}
	}

	counts := []map[string]int{}

	if include_today {
		var count int

		loc, _ := time.LoadLocation("America/Chicago")
		today := time.Now().In(loc)

		err := api.Db.QueryRow(`SELECT total
			FROM daily_counts
			WHERE service_code = $1
				AND ward = $2
				AND requested_date = $3;
			`, service_code, ward_id, today).Scan(&count)

		if err != nil {
			// no rows
			count = 0
		}

		counts = append(counts, map[string]int{today.Format("2006-01-02"): count})
	}

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
		var date time.Time
		var count int

		if err := rows.Scan(&count, &date); err != nil {
			// handle
		}

		counts = append(counts, map[string]int{date.Format("2006-01-02"): count})
	}

	return dumpJson(counts), nil
}

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

	return dumpJson(resp), nil
}

func WardRequestsHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
	// for a given ward, return recent service requests
	vars := mux.Vars(request)
	ward_id := vars["id"]

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

	return dumpJson(result), nil
}

func ServicesHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
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

	return dumpJson(services), nil
}
