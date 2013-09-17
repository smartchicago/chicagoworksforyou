# Data / ETL tools

The CWFY application periodically draws data from external, third-party sources.

This document describes those tools and their prerequisites.

## prerequisites / machine prep

The data tools are written in Python, and rely on a number of modules.  Some of those modules are bundled with the operating system but must be installed separately.  Others require a custom build through easy\_install or pip.

The following commands will prepare the host machine for running the data tools:

	sudo yum -y install pytz python-dateutil numpy scipy python-psycopg2 python-requests
	pip-python install pandas


## configuration

The scripts read configuration details (e.g., database connect string and table names) from an external JSON file.

For example:


	{
	  "db_connect_string" : "host=... dbname=... user=... password=..." ,
	  "db_table_weather_daily_stats" : "weather_daily_stats" ,
	  "db_table_weather_storm_event" : "weather_storm_event"
	}


## the tools

All tools support `--help` for commandline help, and most tools require `--config` to specify the JSON-format configuration file described above.



### load-storm-bulk

We draw historical storm data from the NOAA NCDC Storm Events dataset.  The script `load-storm-bulk` takes care of bulk-loading that data, e.g., to seed a fresh database install.  (A separate tool, described below, is responsible for incremental updates.)

This tool is meant to be run manually, from the commandline.


#### usage

Download historical storm data (described below) and run:

	load-storm-bulk --config config.json file1 file2 ... fileN


This tool loads each file's data as a single transaction.  In the event of a malformed file, then, an operator only needs to correct and reload the errant file.



#### data details


The Storm Events Database updates every 60-90 days, hence, our data will always be two or three months behind the present day.


This dataset is available in different formats:

* the entire dataset (1996-present) is available as a Microsoft Access database file (1GB in size as of this writing, and the Linux `mdbtools` package cannot open it)
* month-by-month data is available from January 2009-present, in individual CSV files, but we are unable to locate a data dictionary for that format. (Some fields use numeric codes instead of descriptive terms like "storm".)
* a different form of CSV data is available by URL query.  This service returns a a maximum 500 results, silently discarding any records beyond that number, so we must query in small enough blocks to not exceed that limit.

We use that last format. Our initial queries (January 2008-August 2013, year-by-year) resulted in fewer than 500 records each, and we expect future (incremental) queries to have fewer than 500 results.

#### fetching historical data

To download data into year-by-year files, provide the proper date specs to the URL query service.  For example:

	for YEAR in $( seq 2008 $( date +'%Y' ) ) ; do
	  curl --output "storm-events-zone-${YEAR}.csv" \
	    "http://www.ncdc.noaa.gov/stormevents/csv?beginDate_mm=01&beginDate_dd=01&beginDate_yyyy=${YEAR}&endDate_mm=12&endDate_dd=31&endDate_yyyy=${YEAR}&eventType=%28Z%29+ALL&county=ALL&zone=COOK&submitbutton=Search&statefips=17%2CILLINOIS"
	done



### load-storm-incremental

The tool `load-storm-incremental` is responsible for updating our storm event data on a regular basis.

It is designed to run as a cron job, though it is also possible to run it from the commandline (e.g., for testing/debugging).

This tool is meant to be run manually, from the commandline.


#### usage

Usage is as follows:

	load-storm-incremental --config config.json [--verbose]


The tool queries the Storm Center service for the date range `(` _most recent entry in the database_ to _today_ `]` , hence it does not require the user to specify the start and end dates to query.

When run with the `--verbose` flag, the tool will print more activity information.  (This is not suitable for cron jobs.)



### load-weather-bulk

For our daily weather statistics -- high/low temperature and precipitation -- we draw data from the NOAA Global Summary of the Day (GSOD) dataset.  The `load-weather-bulk` tool parses historical GSOD data to seed a fresh database.

This tool is meant to be run manually, from the commandline.


#### usage

Download historical weather data (described below) and run:

	load-weather-bulk --config config.json file1 file2 ... fileN

This tool loads each file's data as a single transaction.  In the event of a malformed file, then, an operator only needs to correct and reload the errant file.


#### data details

The GSOD data is provided in a compressed, fixed-width format.  There's one file per year+station, and the current year's file is updated daily.  (That means, to get yesterday's weather stats, we need to download this year's file and scroll to yesterday's date.)

We use the file for the O'Hare Airport weather station, also known as station ID _725300-94846_.


#### fetching historical data

GSOD data is conveniently provided as one file per year.  To fetch historical data for bulk loading, then, one needs to simply fetch each year's file.  For example:

	STATION_ID="725300-94846"
	for YEAR in $( seq 2008 $( date +'%Y' ) ) ; do

	  wget "ftp://ftp.ncdc.noaa.gov/pub/data/gsod/${YEAR}/${STATION_ID}-${YEAR}.op.gz"

	done



### load-weather-incremental

The tool `load-weather-incremental` is responsible for updating our daily weather stats data on a regular basis.

It is designed to run as a cron job, though it is also possible to run it from the commandline (e.g., for testing/debugging).

#### usage

Usage is as follows:

	load-weather-incremental --config config.json [--verbose]

The tool queries the database, to determine the date of the most recent update, and downloads the necessary file(s) from NOAA to update our dataset.



