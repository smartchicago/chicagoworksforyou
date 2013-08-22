package main

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func RequestsMediaHandler(params url.Values, request *http.Request) ([]byte, *ApiError) {
	// Return N most recent days worth of SR that have non-empty media_url values
	//
	// Sample:
	//
	// $ curl "http://localhost:5000/requests/media.json?days=10"
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

	days, err := strconv.Atoi(params.Get("days"))
	if err != nil {
		return nil, &ApiError{Code: 400, Msg: "must specify number of days to return"}
	}

	if days > 30 {
		return nil, &ApiError{Code: 400, Msg: "days must be <= 30"}
	}

	start := time.Now().AddDate(0, 0, -days)

	type SR struct {
		Service_name, Address, Media_url, Service_request_id string
		Ward                                                 int
	}

	var sr_with_media []SR

	rows, err := api.Db.Query(`SELECT service_name,address,media_url,service_request_id,ward 
                FROM service_requests
                WHERE media_url != ''
			AND requested_datetime >= $1
			AND ward IS NOT NULL
                ORDER BY requested_datetime DESC;`, start)

	if err != nil {
		log.Print("error loding media objects ", err)
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
