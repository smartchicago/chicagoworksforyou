package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"io/ioutil"
	"regexp"
	"github.com/kylelemons/go-gypsy/yaml"
	"log"	
)

var (
        dry     =       flag.Bool("dry", false, "dry run: show what will be run but do not execute statements")
        environment = flag.String("environment", "", "Environment to run in, e.g. staging, production")
	config      = flag.String("config", "./config/database.yml", "database configuration file")
	db *sql.DB
)

func init() {	
	flag.Parse()
	
	log.Printf("running in %s environment, configuration file %s", *environment, *config)
	settings := yaml.ConfigFile(*config)

	// setup database connection
	driver, err := settings.Get(fmt.Sprintf("%s.driver", *environment))
	if err != nil {
		log.Fatal("error loading db driver", err)
	}

	connstr, err := settings.Get(fmt.Sprintf("%s.connstr", *environment))
	if err != nil {
		log.Fatal("error loading db connstr", err)
	}

	_db, err := sql.Open(driver, connstr)
	if err != nil {
		log.Fatal("Cannot open database connection", err)
	}

	log.Printf("database connstr: %s", connstr)
	
	db = _db
}

func main() {
	// tool to migrate the database
	// This is very immature code.
	//
	// Usage:
	//
	// go run bin/migrate.go 		# apply migrations
	// go run bin/migrate.go --dry=true 	# show what will be run, do not execute statements.
	//
	// TODO:
	// - rollback
	// - transactional migrations
	// - show output of migration
	// - decent error handling

	// load all migrations and find version numbers
	file_list, err := ioutil.ReadDir("db/")
	if err != nil {
		fmt.Println("error loading directory list", err)
	}

	all_migrations := make(map[string]string)

	re := regexp.MustCompile(`\A(\d+)[^\.]+\.sql\z`)
	for _, item := range file_list {
		if matches := re.FindSubmatch([]byte(item.Name())); matches != nil {
			all_migrations[string(matches[1])] = item.Name()
		}
	}

	fmt.Println("found versions: ", all_migrations)

	// find migrations to apply
	rows, err := db.Query("SELECT * FROM schema_info ORDER BY version DESC;")
	if err != nil {
		fmt.Println("error loading applied migrations", err)
	}

	var applied_migrations []string
	for rows.Next() {
		var version string
		_ = rows.Scan(&version)
		applied_migrations = append(applied_migrations, version)
	}

	fmt.Println("migrations already applied: ", applied_migrations)

	for _, version := range applied_migrations {
		delete(all_migrations, version)
	}

	fmt.Println("migrations that will be applied: ", all_migrations)

	// apply migrations
	for version, migration := range all_migrations {
		// MIGRATE!
		raw_sql, err := ioutil.ReadFile("db/" + migration)
		if err != nil {
			fmt.Printf("error reading %s: %s\n", migration, err)
		}

		fmt.Printf("===== executing migration %s =====\n\n%s\n\n", migration, string(raw_sql))

		if !*dry {
			_, err = db.Exec(string(raw_sql))
			if err != nil {
				fmt.Printf("error migrating %s: %s", migration, err)
				return
			}

			fmt.Println("===== completed migration =====", migration)

			_, err = db.Exec("INSERT INTO schema_info (version) VALUES ($1);", version)
			if err != nil {
				fmt.Printf("error migrating %s: %s", migration, err)
			}
		}
	}
}
