Chicago Works For You API Reference
===================================

Overview
--------

The Chicago Works For You (CWFY) API lives at http://cwfy-api.smartchicagoapps.org/.

The CWFY API serves JSON(P) responses to HTTP requests.

Any request may include a `callback` URL parameter, e.g. `callback=foo`; the response will use the callback parameter as a function name and wrap the response in a Javascript function call.

Health Check
------------

Path: `/health_check`

Description: Display the current status of the system. Returns the current API version, database health, count of service requests in the database, and overall system health.

Input: none

Output:

    $ curl "http://localhost:5000/health_check"
    {
      "Count": 1377257,
      "Database": true,
      "Healthy": true,
      "Version": "0.0.2"
    }

Services
--------

Path: `/services.json`

Description: Return a list of all types of service requests stored in the database and count of each type.

Input: none

Output:

    $ curl "http://localhost:5000/services.json"
    [
      {
        "Count": 354942,
        "Service_code": "4fd3b167e750846744000005",
        "Service_name": "Graffiti Removal"
      },
      {
        "Count": 7970,
        "Service_code": "4fd6e4ece750840569000019",
        "Service_name": "Restaurant Complaint"
      },
      {
        "Count": 92675,
        "Service_code": "4fd3b9bce750846c5300004a",
        "Service_name": "Rodent Baiting / Rat Complaint"
      },
      {
        "Count": 50249,
        "Service_code": "4fd3bbf8e750846c53000069",
        "Service_name": "Tree Debris"
      },
      {
        "Count": 50777,
        "Service_code": "4ffa4c69601827691b000018",
        "Service_name": "Abandoned Vehicle"
      },
      {
        "Count": 27343,
        "Service_code": "4ffa9f2d6018277d400000c8",
        "Service_name": "Street Light 1 / Out"
      },
      {
        "Count": 22489,
        "Service_code": "4ffa971e6018277d4000000b",
        "Service_name": "Pavement Cave-In Survey"
      },
      {
        "Count": 44753,
        "Service_code": "4ffa9cad6018277d4000007b",
        "Service_name": "Alley Light Out"
      },
      {
        "Count": 59801,
        "Service_code": "4fd3bd72e750846c530000cd",
        "Service_name": "Building Violation"
      },
      {
        "Count": 33909,
        "Service_code": "4ffa9db16018277d400000a2",
        "Service_name": "Traffic Signal Out"
      },
      {
        "Count": 4507,
        "Service_code": "4ffa995a6018277d4000003c",
        "Service_name": "Street Cut Complaints"
      },
      {
        "Count": 46949,
        "Service_code": "4fd3b750e750846c5300001d",
        "Service_name": "Sanitation Code Violation"
      },
      {
        "Count": 147457,
        "Service_code": "4fd3b656e750846c53000004",
        "Service_name": "Pothole in Street"
      },
      {
        "Count": 65615,
        "Service_code": "4fd3bd3de750846c530000b9",
        "Service_name": "Street Lights All / Out"
      }
    ]


Time to Close
-------------

Path: `/requests/time_to_close.json`

Description: Return the number of requests closed in a given time interval and the average time to close over the past year. Results are grouped by ward number.

Input:

    end_date: e.g. "2013-06-19"
    count: number of days to count back from end date
    service_code: code of the service, e.g. "4fd3b167e750846744000005"

Output:

The city-wide average time to close and count of requests opened is in the `CityData` object. Time to close is measured in days. The `Threshold` value is one standard deviation below the average number of service requests opened in a ward for the given time period. This value is useful for filtering low-volume wards from the result set.

    $ curl "http://localhost:5000/requests/time_to_close.json?end_date=2013-06-19&count=7&service_code=4fd3b167e750846744000005"
    {
      "WardData": {
         "1": {
             "Time": 6.586492353553241,
              "Count": 643
            },
         ( .. truncated ...)
            "9": {
              "Time": 2.469373385011574,
              "Count": 43
            }
          },
      "CityData": {
        "Time": 3.8197868124884256,
        "Count": 11123
      },
      "Threshold": 27.537741650677532
    }
	

Ward Requests
-------------

Path: `/wards/{id}/requests.json`

Description: Return the 100 most recently updated (by CWFY) requests for a given ward.

Input:

    id: ward number, an integer between 1 - 50, inclusive.

Output:

    $ curl "http://localhost:5000/wards/1/requests.json"
    [
      {
        "Lat": 41.913368,
        "Long": -87.688519,
        "Ward": 1,
        "Police_district": 14,
        "Id": 1378483,
        "Service_request_id": {
          "String": "10-01408091",
          "Valid": true
        },
        "Status": {
          "String": "closed",
          "Valid": true
        },
        "Service_name": {
          "String": "Graffiti Removal",
          "Valid": true
        },
        "Service_code": {
          "String": "4fd3b167e750846744000005",
          "Valid": true
        },
        "Agency_responsible": {
          "String": "Bureau of Street Operations - Graffiti",
          "Valid": true
        },
        "Address": {
          "String": "1743 N ARTESIAN AVE, CHICAGO, IL, 60647",
          "Valid": true
        },
        "Channel": {
          "String": "phone",
          "Valid": true
        },
        "Media_url": {
          "String": "",
          "Valid": true
        },
        "Duplicate": {
          "String": "",
          "Valid": false
        },
        "Parent_service_request_id": {
          "String": "",
          "Valid": false
        },
        "Requested_datetime": {
          "Time": "2010-09-01T13:14:06Z",
          "Valid": true
        },
        "Updated_datetime": {
          "Time": "2010-09-21T18:38:00Z",
          "Valid": true
        },
        "Created_at": {
          "Time": "2013-06-26T22:00:31.628959Z",
          "Valid": true
        },
        "Updated_at": {
          "Time": "2013-06-26T22:00:31.628959Z",
          "Valid": true
        },
        "Extended_attributes": null
      },

      ( result truncated)
    ]


Ward Counts
-----------

Path: `/wards/{id}/counts.json`

Description: Return the number of service requests opened grouped by day, then by service request type, for a given ward.

Input:

    end_date: e.g. "2013-06-19"
    count: number of days to count back from end date
    service_code: code of the service, e.g. "4fd3b167e750846744000005"

Output:

    $ curl "http://localhost:5000/wards/10/counts.json?service_code=4fd3b167e750846744000005&count=7&end_date=2013-06-03"
    {
      "2013-05-28": 10,
      "2013-05-29": 6,
      "2013-05-30": 9,
      "2013-05-31": 3,
      "2013-06-01": 2,
      "2013-06-02": 6,
      "2013-06-03": 7
    }

Request Counts
--------------

Path: `/requests/{service_code}/counts.json`

Description: For a given request service type and date, return the count of requests for that date, grouped by ward, and the city total.

Input:

    end_date: e.g. "2013-06-19"
    count: number of days to count back from end date
    service_code: code of the service, e.g. "4fd3b167e750846744000005"

Output:

The output is a three element map, with keys `DayData`, `CityData`, `WardData`. `DayData` is an array of dates contained in the results. The last element of the array will equal the end_date URL parameter. `CityData` contains the total number of SR opened in the City for the date range (`Count`), and the average number opened per day, for the entire city, over the past 365 days (`Average`). `WardData` contains an array of number of SR opened per day (`Counts`) and average (`Average`) number opened per day over the past 365 days for each of the 50 wards.

    $ curl "http://localhost:5000/requests/4fd3b167e750846744000005/counts.json?end_date=2013-06-19&count=1"
    {
      "DayData": [
        "2013-06-04",
        "2013-06-05",
        "2013-06-06",
        "2013-06-07",
        "2013-06-08",
        "2013-06-09",
        "2013-06-10"
      ],
      "CityData": {
        "Average": 8.084931,
        "Count": 2951
      },
      "WardData": {
        "1": {
          "Counts": [
            29,
            19,
            40,
            60,
            16,
            2,
            35
          ],
          "Average": 16.671232
        },
        "10": {
          "Counts": [
            22,
            2,
            28,
            6,
            2,
            5,
            6
          ],
          "Average": 6.60274
        },
      (response truncated)

Request Counts by Day
---------------------

Path: `/requests/counts_by_day.json`

Description: Given day, return total # of each service type for that day, along with daily average for each service type.

Input:

    day: e.g. 2013-06-20

Output:

    $ curl "http://localhost:5000/requests/counts_by_day.json?day=2013-06-21"
    {
      "4fd3b167e750846744000005": {
        "Count": 379,
        "Average": 8.694054,
        "TopWards": [
          14
        ]
      },
      "4fd3b656e750846c53000004": {
        "Count": 140,
        "Average": 4.195414,
        "TopWards": [
          31
        ]
      },
      "4fd3b750e750846c5300001d": {
        "Count": 82,
        "Average": 1.9250441,
        "TopWards": [
          18
        ]
      },
      "4fd3b9bce750846c5300004a": {
        "Count": 86,
        "Average": 2.774941,
        "TopWards": [
          32,
          50
        ]
      },
      (response truncated)


Requests with media
-------------------

Path: `/requests/media.json`

Description: Return the 500 most recent service requests that have an associated media object.

Input: none

Output:

    $ curl "http://localhost:5000/requests/media.json?callback=foo"
    foo([
      {
        "Service_name": "Pothole in Street",
        "Address": "4552 n Lockwood",
        "Media_url": "http://311request.cityofchicago.org/media/chicago/report/photos/51ded6f0016305b6f8ba12ea/pic_8092_960.png",
        "Service_request_id": "13-00921084",
        "Ward": 45
      },
      {
        "Service_name": "Graffiti Removal",
        "Address": "1545 w Cortez",
        "Media_url": "http://311request.cityofchicago.org/media/chicago/report/photos/51ded522016305b6f8ba12b0/pic_8089_1010.png",
        "Service_request_id": "13-00920959",
        "Ward": 27
      },
      {
        "Service_name": "Sanitation Code Violation",
        "Address": "2031 w 23rd street",
        "Media_url": "http://311request.cityofchicago.org/media/chicago/report/photos/51dece78016305b6f8ba11f8/pic_8086_2356.png",
        "Service_request_id": "13-00920544",
        "Ward": 25
      },
      {
        "Service_name": "Pothole in Street",
        "Address": "4816 - 18 n Linder.",
        "Media_url": "http://311request.cityofchicago.org/media/chicago/report/photos/51decd21016305b6f8ba11ca/pic_8085_960.png",
        "Service_request_id": "13-00920440",
        "Ward": 45
      },
      
Historic Highs
--------------

Path: `/wards/32/historic_highs.json`

Description: For a given ward and service code, return the n-many days with the most service requests opened.

Input:

   	count: 		        number of historical high days to return.
  	service_code:       the code used by the City of Chicago to categorize service requests
   	include_date:       (optional) a YYYY-MM-DD formatted string. If present, the results will include the counts for that day, too. 

Output: 

    $ curl "http://localhost:5000/wards/32/historic_highs.json?service_code=4fd3b167e750846744000005&count=10&include_date=2013-07-25"
    [
      {
        "2013-07-25": 0
      },
      {
        "2010-10-27": 94
      },
      {
        "2008-07-01": 75
      },
      {
        "2010-10-25": 70
      },
      {
        "2008-05-16": 68
      },
      {
        "2010-10-14": 65
      },
      {
        "2008-03-20": 64
      },
      {
        "2009-01-16": 60
      },
      {
        "2008-07-30": 60
      },
      {
        "2008-05-27": 60
      },
      {
        "2008-02-18": 60
      }
    ]