package main

import (
	"flag"
	"fmt"
	"github.com/RangelReale/avistaz-fstop"
	"github.com/RangelReale/filesharetop/importer"
	"gopkg.in/mgo.v2"
	"log"
	"os"
)

var version = flag.Bool("version", false, "show version and exit")
var configfile = flag.String("configfile", "", "configuration file path")

func main() {
	flag.Parse()

	// output version
	if *version {
		fmt.Printf("avistaz-importer version %s\n", fstopimp.VERSION)
		os.Exit(0)
	}

	// create logger
	logger := log.New(os.Stderr, "", log.LstdFlags)

	// load configuration file
	config := avistaz.NewConfig()

	if *configfile != "" {
		logger.Printf("Loading configuration file %s", *configfile)

		err := config.Load(*configfile)
		if err != nil {
			logger.Fatal(err)
		}
	}

	// connect to mongodb
	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Panic(err)
	}
	defer session.Close()

	// create and run importer
	imp := fstopimp.NewImporter(logger, session)
	imp.Database = "fstop_avistaz"

	// create fetcher
	fetcher := avistaz.NewFetcher(config)

	// import data
	err = imp.Import(fetcher)
	if err != nil {
		logger.Fatal(err)
	}

	// consolidate data
	err = imp.Consolidate("", 48)
	if err != nil {
		logger.Fatal(err)
	}

	err = imp.Consolidate("weekly", 168)
	if err != nil {
		logger.Fatal(err)
	}
}
