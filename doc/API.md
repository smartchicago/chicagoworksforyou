Chicago Works For You API Reference
===================================

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
    



router.HandleFunc("/services.json", ServicesHandler)
router.HandleFunc("/requests/time_to_close.json", TimeToCloseHandler)
router.HandleFunc("/wards/{id}/requests.json", WardRequestsHandler)
router.HandleFunc("/wards/{id}/counts.json", WardCountsHandler)
router.HandleFunc("/requests/{service_code}/counts.json", RequestCountsHandler)
router.HandleFunc("/requests/counts_by_day.json", DayCountsHandler)
