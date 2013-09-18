Chicago Works For You API Reference
===================================

Overview
--------

The Chicago Works For You (CWFY) API lives at [http://cwfy-api.smartchicagoapps.org/](http://cwfy-api.smartchicagoapps.org/).

There is a test/staging API running at [http://cwfy-api-staging.smartchicagoapps.org/](http://cwfy-api-staging.smartchicagoapps.org/).

The CWFY API serves JSON(P) responses to HTTP requests. All requests to the API **must** be HTTP GET requests. Sample curl commands are included below. The API does not support any method other than GET.

Requests missing a parameter or with malformed data will get a HTTP 400 response with a JSON representation of the error. HTTP 500 indicates a backend issue and that the request **should not** be retried. The health check endpoint shows the overall health of the system. 

Sample error response:

    $ curl "http://localhost:5000/requests/time_to_close.json?end_date=&count=7&service_code=4fd3b167e750846744000005"
    {
      "message": "invalid end_date",
      "status": 400
    }

Any request may include a `callback` URL parameter, e.g. `callback=foo`; the response will use the callback parameter as a function name and wrap the response in a Javascript function call.

Notes on the data
-----------------

All totals and calculations exclude service requests marked as duplicates.

In some strange cases, the City of Chicago will provide a service request with ward = 0. We save these requests, and in some cases, calling an endpoint with ward = 0 will return these service requests. 

Service requests are fetched from the City of Chicago Open311 API every 30 seconds.

Access/Registration
-------------------

There are no access restrictions to the API at the moment. You do not need to register for access or use a special token to access the API. Smart Chicago appreciates knowing about interesting uses of the API. Developers are encouraged to email [info@smartchicagocollaborative.org](mailto:info@smartchicagocollaborative.org) and share how they're using the API. Smart Chicago reserves the right to block access from applications or users that negatively impact the availability and functionality of the API.

Health Check
------------

Path: `/health_check.json`

Description: Display the current status of the system. Returns the current API version, database health, SR with the greatest 'requested_datetime' field (most recent request), and overall system health. The 'healthy' field indicates overall health, and should be the sole determinate whether or not to use the system.

Input: none

Output:

    $ curl http://localhost:5000/health_check.json
    {
      "most_recent_sr_id": "13-01264162",
      "database": true,
      "healthy": true,
      "version": ""
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
        "count": 123412,
        "service_code": "4fd3bd72e750846c530000cd",
        "service_name": "Building Violation"
      },
      {
        "count": 151704,
        "service_code": "4ffa4c69601827691b000018",
        "service_name": "Abandoned Vehicle"
      },
      {
        "count": 82603,
        "service_code": "4ffa9cad6018277d4000007b",
        "service_name": "Alley Light Out"
      },
      {
        "count": 836103,
        "service_code": "4fd3b167e750846744000005",
        "service_name": "Graffiti Removal"
      },
      {
        "count": 333398,
        "service_code": "4fd3b656e750846c53000004",
        "service_name": "Pothole in Street"
      },
      {
        "count": 132453,
        "service_code": "4fd3bd3de750846c530000b9",
        "service_name": "Street Lights All / Out"
      },
      {
        "count": 39092,
        "service_code": "4ffa971e6018277d4000000b",
        "service_name": "Pavement Cave-In Survey"
      },
      {
        "count": 17815,
        "service_code": "4fd6e4ece750840569000019",
        "service_name": "Restaurant Complaint"
      },
      {
        "count": 189760,
        "service_code": "4fd3b9bce750846c5300004a",
        "service_name": "Rodent Baiting / Rat Complaint"
      },
      {
        "count": 67727,
        "service_code": "4ffa9db16018277d400000a2",
        "service_name": "Traffic Signal Out"
      },
      {
        "count": 7926,
        "service_code": "4ffa995a6018277d4000003c",
        "service_name": "Street Cut Complaints"
      },
      {
        "count": 53229,
        "service_code": "4ffa9f2d6018277d400000c8",
        "service_name": "Street Light 1 / Out"
      },
      {
        "count": 112470,
        "service_code": "4fd3b750e750846c5300001d",
        "service_name": "Sanitation Code Violation"
      },
      {
        "count": 128979,
        "service_code": "4fd3bbf8e750846c53000069",
        "service_name": "Tree Debris"
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
      "ward_data": {
        "1": {
          "time": 2.364411651238426,
          "count": 102
        },
        "10": {
          "time": 4.099598842592593,
          "count": 50
        },
        "11": {
          "time": 0.8528237674768518,
          "count": 151
        },
        "12": {
          "time": 2.1940788901273147,
          "count": 149
        },
        
        (... truncated ...)
        
        "9": {
          "time": 0,
          "count": 0
        }
      },
      "city_data": {
        "time": 2.90123641712963,
        "count": 2587
      },
      "threshold": -3.9964548567631866
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
        "opened": 0,
        "closed": 0
      },
      "2013-08-25": {
        "opened": 0,
        "closed": 0
      },
      "2013-08-26": {
        "opened": 0,
        "closed": 0
      },
      "2013-08-27": {
        "opened": 0,
        "closed": 0
      },
      "2013-08-28": {
        "opened": 0,
        "closed": 0
      },
      "2013-08-29": {
        "opened": 0,
        "closed": 0
      },
      "2013-08-30": {
        "opened": 11,
        "closed": 7
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

The output is a three element map, with keys `DayData`, `CityData`, `WardData`. `DayData` is an array of dates contained in the results. The last element of the array will equal the end_date URL parameter. `CityData` contains the total number of SR opened in the City for the date range (`Count`), and the average number opened per day, for the entire city, over the past 365 days (`Average`). The `DailyMax` field is an array of seven highest counts for SR opened in a day. `WardData` contains an array of number of SR opened per day (`Counts`) and  average (`Average`) number opened per day over the past 365 days for each of the 50 wards.

    $ curl "http://localhost:5000/requests/4fd3b167e750846744000005/counts.json?end_date=2013-06-19&count=1"
    {
      "day_data": [
        "2013-06-19"
      ],
      "city_data": {
        "average": 1.3123288,
        "daily_max": [
          169,
          163,
          161,
          160,
          156,
          155,
          152
        ],
        "count": 479
      },
      "ward_data": {
        "1": {
          "counts": [
            11
          ],
          "average": 16.980822
        },
        "10": {
          "counts": [
            1
          ],
          "average": 6.4054794
        },
        "11": {
          "counts": [
            25
          ],
          "average": 18.536985
        },
        
        (... truncated ...)
        
        "9": {
          "counts": [
            1
          ],
          "average": 0.80547947
        }
      }
    }
    

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
        "count": 387,
        "average": 324.28766,
        "wards": {
          "1": 11,
          "10": 5,
          "11": 7,
          "12": 47,
          
          (... truncated ...)
          
          "8": 0,
          "9": 0
        }
      },
      "4fd3b656e750846c53000004": {
        "count": 140,
        "average": 134.4137,
        "wards": {
          "1": 0,
          "10": 4,
          "11": 2,
          "12": 0,
          
          (... truncated ...)
          
          "8": 2,
          "9": 2
        }
      },
      "4fd3b750e750846c5300001d": {
        "count": 83,
        "average": 43.31781,
        "wards": {
          "1": 2,
          "10": 1,
          "11": 0,
          "12": 2,
          
          (... truncated ...)
          
          "8": 0,
          "9": 2
        }
      },
      "4fd3b9bce750846c5300004a": {
        "count": 86,
        "average": 75.438354,
        "wards": {
          "1": 3,
          "10": 1,
          "11": 0,
          "12": 0,

          (... truncated ...)

          "8": 2,
          "9": 1
        }
      },
      "4fd3bbf8e750846c53000069": {
        "count": 68,
        "average": 51.624657,
        "wards": {
          "1": 1,
          "10": 1,
          "11": 1,
 
          (... truncated ...)

          "8": 2,
          "9": 0
        }
      },
      "4fd3bd3de750846c530000b9": {
        "count": 61,
        "average": 56.09863,
        "wards": {
          "1": 0,
          "10": 0,
          "11": 0,
          "12": 3,
  
          (... truncated ...)

          "8": 6,
          "9": 1
        }
      },
      "4fd3bd72e750846c530000cd": {
        "count": 67,
        "average": 45.673973,
        "wards": {
          "1": 1,
          "10": 5,
          "11": 1,
          "12": 0,
   
          (... truncated ...)

          "7": 0,
          "8": 0,
          "9": 0
        }
      },
      "4fd6e4ece750840569000019": {
        "count": 12,
        "average": 6.893151,
        "wards": {
          "1": 0,
          "10": 0,
          "11": 1,

          (... truncated ...)

          "7": 0,
          "8": 0,
          "9": 0
        }
      },
      "4ffa4c69601827691b000018": {
        "count": 44,
        "average": 38.684933,
        "wards": {
          "1": 2,
          "10": 0,
          "11": 4,
          "12": 1,

          (... truncated ...)

          "7": 1,
          "8": 1,
          "9": 0
        }
      },
      "4ffa971e6018277d4000000b": {
        "count": 23,
        "average": 18.50137,
        "wards": {
          "1": 0,
          "10": 1,
          "11": 0,
 
          (... truncated ...)

          "8": 0,
          "9": 1
        }
      },
      "4ffa995a6018277d4000003c": {
        "count": 3,
        "average": 4.8328767,
        "wards": {
          "1": 0,
          "10": 0,
          "11": 1,
 
          (... truncated ...)

          "8": 0,
          "9": 0
        }
      },
      "4ffa9cad6018277d4000007b": {
        "count": 44,
        "average": 29.767124,
        "wards": {
          "1": 0,
          "10": 3,
          "11": 1,

          (... truncated ...)

          "8": 1,
          "9": 0
        }
      },
      "4ffa9db16018277d400000a2": {
        "count": 23,
        "average": 33.90411,
        "wards": {
          "1": 0,
          "10": 0,
          "11": 0,
          "12": 1,

          (... truncated ...)

          "8": 0,
          "9": 0
        }
      },
      "4ffa9f2d6018277d400000c8": {
        "count": 32,
        "average": 33.2137,
        "wards": {
          "1": 0,
          "10": 1,
          "11": 0,

          (... truncated ...)

          "8": 1,
          "9": 0
        }
      }
    }


Requests with media
-------------------

Path: `/requests/media.json`

Description: Return the 500 most recent service requests that have an associated media object.

Input: none

Output:

    $ curl "http://localhost:5000/requests/media.json?callback=foo"
    foo([
      {
        "service_name": "Graffiti Removal",
        "address": "13536 S Avenue L Arizona",
        "media_url": "http://311request.cityofchicago.org/media/chicago/report/photos/5238e960016302a78310d085/20130909_150213.jpg",
        "service_request_id": "13-01372533",
        "ward": 10
      },
      {
        "service_name": "Graffiti Removal",
        "address": "828-848 North Washtenaw Avenue, Chicago, IL 60622, USA",
        "media_url": "http://311request.cityofchicago.org/media/chicago/report/photos/5238df60016302a78310d04c/1379458553907.jpg",
        "service_request_id": "13-01372401",
        "ward": 26
      },
      {
        "service_name": "Graffiti Removal",
        "address": "Swift Elementary School, 5832 N Winthrop Ave, Chicago, IL  60660",
        "media_url": "http://311request.cityofchicago.org/media/chicago/report/photos/5238daed016302a78310d025/pic_9955_2572.png",
        "service_request_id": "13-01372335",
        "ward": 48
      },

      (... truncated ...)
      
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
        "date": "2010-10-27",
        "count": 94
      },
      {
        "date": "2008-07-01",
        "count": 75
      },

      (... truncated ...)

      {
        "date": "2013-07-25",
        "count": 9
      }
    ]
    
    # If service_code is omitted, all service_codes are returned:
    
    $ curl "http://localhost:5000/wards/2/historic_highs.json?&count=3&include_date=2013-05-23"
    {
      "highs": {
        "4fd3b167e750846744000005": [
          {
            "date": "2008-10-14",
            "count": 55
          },
          {
            "date": "2008-08-07",
            "count": 49
          },
          {
            "date": "2009-05-20",
            "count": 42
          }
        ],
        "4fd3b656e750846c53000004": [
          {
            "date": "2008-01-07",
            "count": 79
          },
          {
            "date": "2008-03-18",
            "count": 62
          },
          {
            "date": "2008-03-27",
            "count": 53
          }
        ],

        (... truncated ...)

        "4ffa9f2d6018277d400000c8": [
          {
            "date": "2013-02-08",
            "count": 41
          },
          {
            "date": "2012-05-21",
            "count": 23
          },
          {
            "date": "2012-05-16",
            "count": 22
          }
        ]
      },
      "current": {
        "4fd3b167e750846744000005": {
          "date": "2013-05-23",
          "count": 8
        },
   
        (... truncated ...)

        "4ffa9f2d6018277d400000c8": {
          "date": "2013-05-23",
          "count": 0
        }
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
      "incoming": [
        {
          "id": 124,
          "ward_2013": 39,
          "ward_2015": 50,
          "boundary": "{\"type\":\"MultiPolygon\",\"crs\":{\"type\":\"name\",\"properties\":{\"name\":\"EPSG:4326\"}},\"coordinates\":[[[[-87.72275,41.99706],[-87.72275,41.99705],[-87.72275,41.99693],[-87.72275,41.99706]]],[[[-87.72275,41.99677],[-87.72274,41.99671],[-87.72274,41.99652],[-87.72273,41.99633],[-87.72273,41.99612],[-87.72273,41.99593],[-87.72272,41.99577],[-87.72272,41.99549],[-87.72272,41.99549],[-87.72272,41.99549],[-87.72273,41.99585],[-87.72275,41.99677]]],[[[-87.72144,41.99373],[-87.72144,41.99373],[-87.72144,41.99359],[-87.72144,41.99373]]],[[[-87.72142,41.99293],[-87.72142,41.99283],[-87.72142,41.99282],[-87.72142,41.99293]]],[[[-87.72139,41.99148],[-87.72138,41.99133],[-87.72138,41.99126],[-87.72138,41.99112],[-87.72137,41.99094],[-87.72137,41.99077],[-87.72137,41.99066],[-87.72137,41.99063],[-87.72137,41.99066],[-87.72139,41.99148]]],[[[-87.72136,41.99023],[-87.72102,41.99023],[-87.72048,41.99024],[-87.72011,41.99025],[-87.71946,41.99025],[-87.71919,41.99026],[-87.71901,41.99026],[-87.71911,41.99026],[-87.71986,41.99025],[-87.72011,41.99025],[-87.72036,41.99024],[-87.72111,41.99023],[-87.72136,41.99023],[-87.72136,41.99023]]],[[[-87.71877,41.99026],[-87.7182,41.99027],[-87.71793,41.99027],[-87.71862,41.99027],[-87.71877,41.99026]]],[[[-87.71678,41.99029],[-87.71656,41.99029],[-87.71667,41.99029],[-87.71678,41.99029]]],[[[-87.71628,41.9903],[-87.71569,41.9903],[-87.71525,41.99031],[-87.71545,41.99031],[-87.71618,41.9903],[-87.71628,41.9903]]],[[[-87.71511,41.99031],[-87.71455,41.99032],[-87.71444,41.99032],[-87.71408,41.99032],[-87.71423,41.99032],[-87.71444,41.99032],[-87.71455,41.99032],[-87.71496,41.99031],[-87.71511,41.99031]]],[[[-87.71389,41.99032],[-87.71373,41.99033],[-87.71341,41.99033],[-87.71318,41.99033],[-87.71276,41.99034],[-87.7127,41.99034],[-87.71276,41.99034],[-87.71301,41.99034],[-87.71374,41.99033],[-87.71389,41.99032]]],[[[-87.71156,41.99036],[-87.71143,41.99036],[-87.71102,41.99036],[-87.71137,41.99036],[-87.71156,41.99036]]],[[[-87.71102,41.99036],[-87.7108,41.99036],[-87.7109,41.99036],[-87.71102,41.99036]]],[[[-87.71007,41.99037],[-87.70979,41.99038],[-87.70969,41.99038],[-87.71007,41.99037]]],[[[-87.70945,41.99038],[-87.70931,41.99038],[-87.70942,41.99038],[-87.70945,41.99038]]],[[[-87.70914,41.99038],[-87.70909,41.99038],[-87.70909,41.99038],[-87.70913,41.99038],[-87.70914,41.99038]]]]}"
        },

        (... truncated ...)

        }
      ],
      "outgoing": [
        {
          "id": 189,
          "ward_2013": 50,
          "ward_2015": 39,
          "boundary": "{\"type\":\"MultiPolygon\",\"crs\":{\"type\":\"name\",\"properties\":{\"name\":\"EPSG:4326\"}},\"coordinates\":[[[[-87.70916,41.99038],[-87.70931,41.99038],[-87.7093,41.99038],[-87.70916,41.99038],[-87.70916,41.99038]]],[[[-87.70945,41.99038],[-87.70963,41.99038],[-87.70969,41.99038],[-87.70948,41.99038],[-87.70945,41.99038]]],[[[-87.71007,41.99037],[-87.71013,41.99037],[-87.71051,41.99037],[-87.7108,41.99036],[-87.71057,41.99037],[-87.71045,41.99037],[-87.71019,41.99037],[-87.71007,41.99037]]],[[[-87.71156,41.99036],[-87.71164,41.99036],[-87.71169,41.99035],[-87.71196,41.99035],[-87.71235,41.99034],[-87.7127,41.99034],[-87.71241,41.99034],[-87.71168,41.99036],[-87.71156,41.99036]]],[[[-87.71389,41.99032],[-87.71398,41.99032],[-87.71408,41.99032],[-87.71398,41.99032],[-87.71389,41.99032]]],[[[-87.71511,41.99031],[-87.7152,41.99031],[-87.71525,41.99031],[-87.7152,41.99031],[-87.71511,41.99031]]],[[[-87.71628,41.9903],[-87.71642,41.99029],[-87.71656,41.99029],[-87.71642,41.99029],[-87.71628,41.9903]]],[[[-87.71678,41.99029],[-87.71692,41.99029],[-87.71764,41.99028],[-87.71793,41.99027],[-87.71789,41.99028],[-87.71764,41.99028],[-87.7174,41.99028],[-87.71678,41.99029]]],[[[-87.71877,41.99026],[-87.71886,41.99026],[-87.71901,41.99026],[-87.71886,41.99026],[-87.71877,41.99026]]],[[[-87.72136,41.99023],[-87.72136,41.99023],[-87.72137,41.99063],[-87.72136,41.99057],[-87.72136,41.99023]]],[[[-87.72139,41.99148],[-87.7214,41.99196],[-87.7214,41.99216],[-87.72141,41.99236],[-87.72142,41.99282],[-87.72141,41.99231],[-87.7214,41.99196],[-87.72139,41.99161],[-87.72139,41.99148]]],[[[-87.72142,41.99293],[-87.72142,41.99306],[-87.72143,41.99327],[-87.72144,41.99359],[-87.72143,41.99337],[-87.72142,41.99293]]],[[[-87.72144,41.99373],[-87.72167,41.99373],[-87.72199,41.99372],[-87.72206,41.99372],[-87.72234,41.99372],[-87.72268,41.99372],[-87.72269,41.99408],[-87.72269,41.99432],[-87.7227,41.99455],[-87.7227,41.99476],[-87.72271,41.99497],[-87.72271,41.99518],[-87.72272,41.99549],[-87.72247,41.99549],[-87.7221,41.9955],[-87.72173,41.9955],[-87.72148,41.9955],[-87.72147,41.99514],[-87.72145,41.99408],[-87.72144,41.99373],[-87.72144,41.99373]]],[[[-87.72275,41.99677],[-87.72275,41.99679],[-87.72275,41.99686],[-87.72275,41.99693],[-87.72275,41.99691],[-87.72275,41.99679],[-87.72275,41.99677]]],[[[-87.72275,41.99706],[-87.72279,41.99726],[-87.72279,41.99726],[-87.72276,41.99716],[-87.72275,41.99706]]],[[[-87.71102,41.99036],[-87.71102,41.99036],[-87.71102,41.99036],[-87.71102,41.99036],[-87.71102,41.99036]]]]}"
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
      "time": 38.26763888888889,
      "count": 3
    }