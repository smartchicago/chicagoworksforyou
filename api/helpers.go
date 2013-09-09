package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func WrapJson(unwrapped []byte, callback []string) (jsn []byte) {
	jsn = unwrapped
	if len(callback) > 0 {
		wrapped := strings.Join([]string{callback[0], "(", string(jsn), ");"}, "")
		jsn = []byte(wrapped)
	}

	return
}

func endpoint(f ApiEndpoint) http.HandlerFunc {
	// define an endpoint and setup consistent logging

	return func(w http.ResponseWriter, req *http.Request) {
		w = setHeaders(w)
		params := req.URL.Query()

		log.Printf("[cwfy %s] %s %s%s %+v", version, req.RemoteAddr, req.Host, req.RequestURI, params)

		t := time.Now()
		response, err := f(params, req)

		if err != nil {
			log.Printf(err.Error())
			http.Error(w, string(dumpJson(err)), err.Code)
		}

		w.Write(WrapJson(response, params["callback"]))
		diff := time.Now()
		log.Printf("[cwfy %s] %s %s%s completed in %v", version, req.RemoteAddr, req.Host, req.RequestURI, diff.Sub(t))
	}
}

func setHeaders(w http.ResponseWriter) http.ResponseWriter {
	// set HTTP headers on the response object
	// TODO: add cache control headers

	w.Header().Set("Content-type", "application/json; charset=utf-8")
	w.Header().Set("Server", fmt.Sprintf("ChicagoWorksForYou.com/%s", version))
	return w
}

func dumpJson(in interface{}) []byte {
	out, err := json.MarshalIndent(in, "", "  ")
	if err != nil {
		log.Printf("error marshalling to json: %s", err)
	}
	return out
}

func backend_error(err error) ([]byte, *ApiError) {
	log.Printf("backend error: %s", err)
	return nil, &ApiError{Msg: "backend error, retry your request later.", Code: 500}
}
