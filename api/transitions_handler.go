package main

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
)

func TransitionsHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
	// return the transition areas

	ward := params.Get("ward")

	type Transition struct {
		Id, Ward2013, Ward2015 int
		Boundary               string
	}

	type Changes struct {
		Incoming, Outgoing []Transition
	}

	var c Changes

	w, err := strconv.Atoi(ward)
	if err != nil || w < 1 || w > 50 {
		return nil, &ApiError{Code: 400, Msg: "invalid ward"}
	}

	rows, err := api.Db.Query(`SELECT ward_2013,ward_2015,id, ST_AsGeoJSON(boundary, 5, 2) 
		FROM transition_areas 
		WHERE ward_2013 = $1 
			OR ward_2015 = $2 
		ORDER BY ward_2013;`, w, w)
	if err != nil {
		log.Printf("error fetching transition areas: %s", err)
	}

	for rows.Next() {
		var t Transition
		if err := rows.Scan(&t.Ward2013, &t.Ward2015, &t.Id, &t.Boundary); err != nil {
			log.Printf("error loading transition area result: %s", err)
		}

		if t.Ward2013 == w {
			c.Outgoing = append(c.Outgoing, t)
		} else {
			c.Incoming = append(c.Incoming, t)
		}
	}

	return dumpJson(c), nil
}
