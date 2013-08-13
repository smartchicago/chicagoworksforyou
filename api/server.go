package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
)

type Api struct {
	Db *sql.DB
}

var (
	api          Api
	env          Environment
	version      string // set at compile time, will be the current git hash
	environment  = flag.String("environment", "", "Environment to run in, e.g. staging, production")
	config       = flag.String("config", "./config/database.yml", "database configuration file")
	port         = flag.Int("port", 5000, "port that server will listen to (default: 5000)")
	ServiceCodes = []string{"4fd3bd72e750846c530000cd", "4ffa9cad6018277d4000007b", "4ffa4c69601827691b000018", "4fd3b167e750846744000005", "4fd3b656e750846c53000004", "4ffa971e6018277d4000000b", "4fd3bd3de750846c530000b9", "4fd6e4ece750840569000019", "4fd3b9bce750846c5300004a", "4ffa9db16018277d400000a2", "4ffa995a6018277d4000003c", "4fd3bbf8e750846c53000069", "4fd3b750e750846c5300001d", "4ffa9f2d6018277d400000c8"}
)

func init() {
	log.Printf("starting ChicagoWorksforYou.com API server version %s", version)

	// load db config
	flag.Parse()
	log.Printf("running in %s environment, configuration file %s", *environment, *config)

	api.Db = env.Load(config, environment)
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
