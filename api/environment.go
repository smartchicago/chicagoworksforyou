package main

import (
"github.com/kylelemons/go-gypsy/yaml"
"database/sql"
"fmt"
"log"
)

type Environment struct {
	Name, Driver, Connstr string
}

func (env *Environment) Load(config, name *string) *sql.DB {
	settings := yaml.ConfigFile(*config)

	// setup database connection
	driver, err := settings.Get(fmt.Sprintf("%s.driver", *name))	
	if err != nil {
		log.Fatal("error loading db driver", err)
	}
	env.Driver = driver
		
	connstr, err := settings.Get(fmt.Sprintf("%s.connstr", *name))
	if err != nil {
		log.Fatal("error loading db connstr", err)
	}
	env.Connstr = connstr

	db, err := sql.Open(env.Driver, env.Connstr)
	if err != nil {
		log.Fatal("Cannot open database connection", err)
	}

	log.Printf("database connstr: %s", env.Connstr)
	
	return db
}


