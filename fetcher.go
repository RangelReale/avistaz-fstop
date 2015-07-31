package avistaz

import (
	"github.com/RangelReale/filesharetop/lib"
	"io/ioutil"
	"log"
)

type Fetcher struct {
	logger *log.Logger
	config *Config
}

func NewFetcher(config *Config) *Fetcher {
	return &Fetcher{
		logger: log.New(ioutil.Discard, "", 0),
		config: config,
	}
}

func (f *Fetcher) ID() string {
	return "AVISTAZ"
}

func (f *Fetcher) SetLogger(l *log.Logger) {
	f.logger = l
}

func (f *Fetcher) Fetch() (map[string]*fstoplib.Item, error) {
	parser := NewAVParser(f.config, f.logger)

	// parse 4 pages ordered by seeders
	err := parser.Parse(AVSORT_SEEDERS, AVSORTBY_DESCENDING, 4)
	if err != nil {
		return nil, err
	}

	// parse 2 pages ordered by leechers
	err = parser.Parse(AVSORT_LEECHERS, AVSORTBY_DESCENDING, 2)
	if err != nil {
		return nil, err
	}

	return parser.List, nil
}

func (f *Fetcher) CategoryMap() (*fstoplib.CategoryMap, error) {
	return &fstoplib.CategoryMap{
		"MOVIE": []string{"film"},
		"TV":    []string{"tv"},
		"MUSIC": []string{"music"},
		"CAT3":  []string{"cat3"},
	}, nil
}
