package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

func WardBoundariesHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
	// Given a latitude and longitude, return the 2013 and 2015 wards that contain it.
	//
	// Example:
	//
	// $ curl "http://localhost:5000/wards/boundaries.json?lat=41.8710&long=-87.6227"
	// {
	//   "2013": 2,
	//   "2015": 4
	// }

	lat, err := strconv.ParseFloat(params["lat"][0], 32)
	if err != nil {
		return nil, &ApiError{Code: 400, Msg: "bad latitude value"}
	}

	long, err := strconv.ParseFloat(params["long"][0], 32)
	if err != nil {
		return nil, &ApiError{Code: 400, Msg: "bad longitude value"}
	}

	type WardLocation struct {
		Ward_2013 int `json:"2013"`
		Ward_2015 int `json:"2015"`
	}

	var wl WardLocation

	query := fmt.Sprintf("SELECT ward FROM ward_boundaries_2013 WHERE ST_Contains(boundary, ST_PointFromText('POINT(%f %f)', 4326))", long, lat)

	err = api.Db.QueryRow(query).Scan(&wl.Ward_2013)
	if err != nil {
		log.Print(err)
	}

	query = fmt.Sprintf("SELECT ward FROM ward_boundaries_2015 WHERE ST_Contains(boundary, ST_PointFromText('POINT(%f %f)', 4326))", long, lat)

	err = api.Db.QueryRow(query).Scan(&wl.Ward_2015)
	if err != nil {
		log.Print(err)
	}

	return dumpJson(wl), nil
}
