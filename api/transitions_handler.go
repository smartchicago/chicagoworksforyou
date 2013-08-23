package main

import (
	"log"
	"net/http"
	"net/url"
	// "strconv"
	// "time"
)

func TransitionsHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
	// return the transition areas

	ward := params.Get("ward")
	
	type Transition struct {
		Id, Ward2013, Ward2015 int
		Boundary string
	}
	
	var transitions []Transition
	// [ { 'id': 123, 'Ward2013': 42, 'Ward2015': 35, 'Boundary': <GeoJSON> },  ] }	
	
	rows, err := api.Db.Query(`SELECT ward_2013,ward_2015,id, ST_AsGeoJSON(boundary) FROM transition_areas ORDER BY ward_2013;`)
	if err != nil {
		log.Printf("error fetching transition areas: %s", err)
	}
	
	for rows.Next() {
		var t Transition
		if err := rows.Scan(&t.Ward2013, &t.Ward2015, &t.Id, &t.Boundary); err != nil {
			log.Printf("error loading transition area result: %s", err)
		}
		transitions = append(transitions, t)
	}		

	return dumpJson(transitions), nil
}