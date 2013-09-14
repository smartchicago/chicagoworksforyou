Chicago Works For You API Reference
===================================

Overview
--------

The Chicago Works For You (CWFY) API lives at http://cwfy-api.smartchicagoapps.org/.

There is a test/staging API running at http://cwfy-api-staging.smartchicagoapps.org/.

The CWFY API serves JSON(P) responses to HTTP requests. All requests to the API **must** be HTTP GET requests. Sample curl commands are included below. The API does not support any method other than GET.

Requests missing a parameter or with malformed data will get a HTTP 400 response with a error message in the body. HTTP 500 indicates a backend issue and that the request **should not** be retried. The health check endpoint shows the overall health of the system. 

Any request may include a `callback` URL parameter, e.g. `callback=foo`; the response will use the callback parameter as a function name and wrap the response in a Javascript function call.

Access/Registration
-------------------

There are no access restrictions to the API at the moment. You do not need to register for access or use a special token to access the API. Smart Chicago appreciates knowing about interesting uses of the API. Developers are encouraged to email info@smartchicagocollaborative.org and share how they're using the API. Smart Chicago reserves the right to block access from applications or users that negatively impact the availability and functionality of the API.

Health Check
------------

Path: `/health_check`

Description: Display the current status of the system. Returns the current API version, database health, SR with the greatest 'requested_datetime' field (most recent request), and overall system health. The 'healthy' field indicates overall health, and should be the sole determinate whether or not to use the system.

Input: none

Output:

    $ curl http://localhost:5000/health_check
    {
      "most_recent_sr_id": "13-01255471",
      "Database": true,
      "Healthy": true,
      "Version": ""
    }

Services
--------

Path: `/services.json`

Description: Return a list of all types of service requests stored in the database and count of each type. The database contains all Chicago service request data going back to Jan 1, 2008.

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
    service_code: (optional) code of the service, e.g. "4fd3b167e750846744000005"

Output:

The city-wide average time to close and count of requests opened is in the `CityData` object. Time to close is measured in days. The `Threshold` value is one standard deviation below the average number of service requests opened in a ward for the given time period. This value is useful for filtering low-volume wards from the result set. If `service_code` is omitted, the values will be for all service types.

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
	

Ward Counts
-----------

Path: `/wards/{id}/counts.json`

Description: Return the number of service requests opened and closed grouped by day, then by service request type, for a given ward. If `service_code` is omitted, return a count of all SR opened and closed each day.

Input:

    end_date: e.g. "2013-06-19"
    count: number of days to count back from end date
    service_code: (optional) code of the service, e.g. "4fd3b167e750846744000005"

Output:

    $ curl "http://localhost:5000/wards/10/counts.json?count=7&end_date=2013-08-30"
    {
      "2013-08-24": {
        "Opened": 0,
        "Closed": 0
      },
      "2013-08-25": {
        "Opened": 0,
        "Closed": 0
      },
      "2013-08-26": {
        "Opened": 7,
        "Closed": 4
      },
      "2013-08-27": {
        "Opened": 20,
        "Closed": 37
      },
      "2013-08-28": {
        "Opened": 18,
        "Closed": 34
      },
      "2013-08-29": {
        "Opened": 7,
        "Closed": 6
      },
      "2013-08-30": {
        "Opened": 0,
        "Closed": 0
      }
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

Path: `/wards/{ward}/historic_highs.json`

Description: For a given ward and service code, return the n-many days with the most service requests opened.

Input:

   	count: 		        number of historical high days to return.
  	service_code:       (optional) the code used by the City of Chicago to categorize service requests. If omitted, all service types will be returned
   	include_date:       a YYYY-MM-DD formatted string. If present, the results will include the counts for that day, too. 

Output: 

    $ curl "http://localhost:5000/wards/32/historic_highs.json?service_code=4fd3b167e750846744000005&count=10&include_date=2013-07-25"
    [
      {
        "Date": "2013-04-30",
        "Count": 37
      },
      {
        "Date": "2013-05-20",
        "Count": 36
      },
      {
        "Date": "2013-04-21",
        "Count": 31
      },
      (... truncated ...)
    ]
    
    # If service_code is omitted, all service_codes are returned:
    
    $ curl "http://localhost:5000/wards/2/historic_highs.json?&count=3&include_date=2013-05-23"
    {
      "Highs": {
        "4fd3b167e750846744000005": [
          {
            "Date": "2013-06-27",
            "Count": 35
          },
          {
            "Date": "2013-05-07",
            "Count": 32
          },
          {
            "Date": "2013-05-20",
            "Count": 27
          }
        ],
        "4fd3b656e750846c53000004": [
          {
            "Date": "2013-06-28",
            "Count": 23
          },
          {
            "Date": "2013-07-09",
            "Count": 19
          },
          {
            "Date": "2013-08-05",
            "Count": 12
          }
        ],
        (... truncated ...)

        ]
      },
      "Current": {
        "4fd3b167e750846744000005": {
          "Date": "2013-05-23",
          "Count": 8
        },
        "4fd3b656e750846c53000004": {
          "Date": "2013-05-23",
          "Count": 3
        },
        "4fd3b750e750846c5300001d": {
          "Date": "2013-05-23",
          "Count": 0
        },
        (... truncated ...)
 
      }
    }
    
Ward Transitions
----------------

Path:   `/wards/transitions.json`

Description: Chicago ward boundaries are changing in 2015. This endpoint returns a list of all areas in the city that are changing from one ward to another. The response includes the current ward, 2015 ward, unique ID of the transition area, and a GeoJSON representation of the area. Transition areas are organized into "incoming" and "outgoing" groups.

Input: 

        ward: integer 1..50

Output: 

    $ curl "http://localhost:5000/wards/transitions.json?ward=50"

    {
      "Incoming": [
        {
          "Id": 37,
          "Ward2013": 26,
          "Ward2015": 1,
          "Boundary": (... truncated geojson ...) 
        },
        {
          "Id": 182,
          "Ward2013": 27,
          "Ward2015": 1,
          "Boundary": (... truncated geojson ...) 
        },
        {
          "Id": 69,
          "Ward2013": 32,
          "Ward2015": 1,
          "Boundary": (... truncated geojson ...) 
        },
        {
          "Id": 59,
          "Ward2013": 35,
          "Ward2015": 1,
          "Boundary": (... truncated geojson ...) 
        }
      ],
      "Outgoing": [
        {
          "Id": 103,
          "Ward2013": 1,
          "Ward2015": 26,
          "Boundary": (... truncated geojson ...) 
        },
        {
          "Id": 105,
          "Ward2013": 1,
          "Ward2015": 33,
          "Boundary": (... truncated geojson ...) 
        },
        {
          "Id": 104,
          "Ward2013": 1,
          "Ward2015": 32,
          "Boundary": (... truncated geojson ...) 
        },
        (... truncated ...) 
      ]
    }
    
Transition Time To Close
------------------------

Path:   `/transitions/time_to_close.json`

Description: Calculate the average time to close for service requests in a given transition area.

Input:

    transition_area_id: integer ID of the transition area. Required.
    service_code: (optional) limit the TTC average to a given service code. If omitted, all service types will be averaged.
    count: number of days to go back in time
    end_date: date (YYYY-MM-DD) to base calculations from
    
Output:

    $ curl "http://localhost:5000/transitions/time_to_close.json?transition_area_id=1&count=7&end_date=2013-08-22"
    {
      "Time": 0.04724537037037037,
      "Count": 1
    }