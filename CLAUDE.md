# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`avistaz-fstop` is a site plugin for [`github.com/RangelReale/filesharetop`](https://github.com/RangelReale/filesharetop), the shared core that handles MongoDB storage, scoring and the web UI. This repo only contains the scraper for https://avistaz.to/torrents?... and two thin `main` binaries. Almost all behavior beyond HTML parsing lives in the core packages `fstoplib` (`lib/`), `fstopimp` (`importer/`) and `fstopsite` (`site/`). Read the core repo's CLAUDE.md for the data model and scoring.

This is legacy, pre-modules Go code. There is no `go.mod`, and imports assume GOPATH layout (`$GOPATH/src/github.com/RangelReale/avistaz-fstop`, with `filesharetop` next to it). Dependencies include `github.com/PuerkitoBio/goquery`, `gopkg.in/mgo.v2`, `github.com/BurntSushi/toml` and, through the core site package, `code.google.com/p/plotinum`.

## Layout

- Root package `avistaz`: `fetcher.go` implements `fstoplib.Fetcher` (`ID`, `SetLogger`, `Fetch`, `CategoryMap`). `avparser.go` (`AVParser`) downloads pages and parses them with goquery into `map[id]*fstoplib.Item`.
- `avistaz-importer/`: run periodically (hourly). It fetches, calls `Importer.Import`, then `Consolidate("", 48)` and `Consolidate("weekly", 168)` into database `fstop_avistaz`.
- `avistaz-site/`: runs `fstopsite.RunServer` on port **13113** against `fstop_avistaz`, with `TopId = "weekly"`.
- Both binaries connect to MongoDB at `localhost` (hard-coded) and support `-version`.

## Scraping

It scrapes the global torrent list: 4 pages sorted by seeders, then 2 sorted by leechers. The parser takes each row's category from its Font Awesome icon class (`fa-film`, `fa-tv`, …; a `text-pink` icon means `cat3`).

When editing the parser:
- Create items with `fstoplib.NewItem()` so that unavailable stats stay `-1`.
- Dedupe on the item ID. The same torrent can appear in both the seeders pass and the leechers pass. `SeedersPos` / `LeechersPos` record its rank in each pass.
- `AddDate` must be `YYYY-MM-DD`.
- `Link` must be an absolute URL.
- Rows that fail to parse are logged and skipped, not treated as errors.

## Configuration

`avistaz.conf` (TOML) holds the login session cookies (`avistaz_session`, `XSRF_TOKEN`, `Avistazlove`, `Cfduid`, `Remember_id`, `Remember`). The parser sends them on every request. The importer takes `-configfile <path>`. Put real values in `avistaz-current.conf`, which is gitignored, not in the committed template.

## Commands

```sh
go build ./avistaz-importer ./avistaz-site
go test -run TestFetcher .     # single test
```

`TestFetcher` hits the live site and needs valid cookies. `TestDownload` is disabled (it starts with `return`).
