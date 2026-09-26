# avistaz-fstop

[filesharetop](https://github.com/RangelReale/filesharetop) plugin for AvistaZ (avistaz.to). It scrapes the tracker's top torrent lists and serves a ranked "Avistaz Top" website.

## Components

- **`avistaz-importer`** scrapes the tracker and stores a snapshot in MongoDB (database `fstop_avistaz`). It then recomputes the 48-hour and weekly rankings. Run it periodically, for example hourly.
- **`avistaz-site`** is the web UI, served on port **13113**.

Both binaries connect to MongoDB on the local host and accept `-version`.

## Configuration

AvistaZ requires a login. Copy `avistaz.conf` to `avistaz-current.conf` (this copy is gitignored) and fill in the session cookies from a logged-in browser: `avistaz_session`, `XSRF_TOKEN`, `Avistazlove`, `Cfduid`, `Remember_id`, `Remember`. Pass it with `-configfile avistaz-current.conf`.

## Build

This is pre-modules Go code. Put this repo and `filesharetop` in a GOPATH workspace, then build:

```sh
go get github.com/RangelReale/avistaz-fstop/...
# or, from a GOPATH checkout:
go build ./avistaz-importer ./avistaz-site
```

## Run

```sh
# once per hour, e.g. from cron
./avistaz-importer/avistaz-importer -configfile avistaz-current.conf

# web UI at http://localhost:13113/
./avistaz-site/avistaz-site
```

## Tests

```sh
go test -run TestFetcher .
```

The test scrapes the live tracker, so it needs network access. It runs with an empty config, so it gets results only if the tracker serves the list without logging in.
