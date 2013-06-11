package main

import (
        "log"
        "github.com/gorilla/mux"
        "net/http"
        "encoding/json"
)

func main() {
        log.Print("starting ChicagoWorksforYou.com API server")        
        router := mux.NewRouter()
        router.HandleFunc("/health_check", HealthCheckHandler)
        http.ListenAndServe(":4000", router)
}

func HealthCheckHandler(response http.ResponseWriter, request *http.Request) {
        response.Header().Add("Content-type", "application/json")        
        health_check := map[string]string{}
        health_check["database"] = "dbconn"
        health_check["sr_count"] = "123"
        jsn, _ := json.Marshal(health_check)
        response.Write(jsn)
}
