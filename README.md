Chicago Works For You
=====================

![CWFY screenshot](doc/cwfy-screenshot.png)


Chicago Works For You ([http://www.chicagoworksforyou.com](http://www.chicagoworksforyou.com)) is a citywide dashboard with ward-by-ward views of service delivery in Chicago. 

Technical
---------

The CWFY website ([`frontend/`](frontend/)) runs on Jekyll and Compass. There is a backend API ([`api/`](api/)) written in Go.

See [API.md](doc/API.md) for details on how to use the backend API. See [FRONTEND.md](doc/FRONTEND.md) for information on the public-facing website.

Data
----

The service request data for Chicago Works For You comes from the [City of Chicago Open311 API](http://dev.cityofchicago.org/docs/api).

Chicago Works For You publishes a nightly database snapshot. This is a complete copy of the production database powering [chicagoworksforyou.com](http://chicagoworksforyou.com).

[Download the snapshot](http://chicagoworksforyou.com/about/#can_i_use_your_data)

Instructions for loading into a local PostgreSQL database:

    createdb cwfy
    pg_restore -d cwfy -O -c /path/to/download/production.dump

This assumes that your have PostgreSQL installed and the PostGIS extension installed. If you are using Mac OS X,  [Postgres.app](http://postgresapp.com/) is a very quick and easy way to install both PostgreSQL and PostGIS.
The database schema is available at [db/schema.sql](db/schema.sql).

Contributing
------------

We welcome contributions to the application. A few guidelines:

 * Fork this repository
 * Create a [topic branch](http://git-scm.com/book/en/Git-Branching-Branching-Workflows#Topic-Branches)
 * Open a pull request with a concise description of the change. Bonus points for a screenshot.

License
-------

The application code is released under the [MIT License](LICENSE.md). Editorial content is released under the Creative Commons [Attribution 3.0 Unported (CC BY 3.0)](http://creativecommons.org/licenses/by/3.0/deed.en_US) license. Content from other authors (e.g. photos on service types pages) are used according to their licenses.

Credits
-------

CWFY was built by [Daniel X. O'Neil](https://github.com/danxoneil), [Sandor Weisz](https://github.com/santheo), and [Chris Gansen](https://github.com/cgansen).

We thank [Christopher Whitaker](https://github.com/govintrenches), [Q Ethan McCallum](https://github.com/qethanm), [Rob Brackett](https://github.com/mr0grog), [Eryan Cobham](https://github.com/littlelazer), and [Paul Smith](https://github.com/paulsmith) for valuable contributions and feedback.
