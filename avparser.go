package avistaz

import (
	"errors"
	"fmt"
	gq "github.com/PuerkitoBio/goquery"
	"github.com/RangelReale/filesharetop/lib"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type AVSort string

const (
	AVSORT_SEEDERS  AVSort = "seed"
	AVSORT_LEECHERS AVSort = "leech"
	AVSORT_COMPLETE AVSort = "complete"
)

type AVSortBy string

const (
	AVSORTBY_ASCENDING  AVSortBy = "asc"
	AVSORTBY_DESCENDING AVSortBy = "desc"
)

type AVParser struct {
	List   map[string]*fstoplib.Item
	config *Config
	logger *log.Logger
}

func NewAVParser(config *Config, l *log.Logger) *AVParser {
	return &AVParser{
		List:   make(map[string]*fstoplib.Item),
		config: config,
		logger: l,
	}
}

func (p *AVParser) Parse(sort AVSort, sortby AVSortBy, pages int) error {

	if pages < 1 {
		return errors.New("Pages must be at least 1")
	}

	posct := int32(0)
	for pg := 1; pg <= pages; pg++ {
		var doc *gq.Document
		var e error

		// download the page
		u, e := url.Parse(fmt.Sprintf("https://avistaz.to/torrents?&order=%s&sort=%s&page=%d", sort, sortby, pg))
		if e != nil {
			return e
		}

		cookies, _ := cookiejar.New(nil)
		cookies.SetCookies(u, []*http.Cookie{
			&http.Cookie{Name: "avistaz_session", Value: p.config.Avistaz_session, Path: "/", Domain: "avistaz.to"},
		})

		client := &http.Client{
			Jar: cookies,
		}

		req, e := http.NewRequest("GET", u.String(), nil)
		if e != nil {
			return e
		}
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.3; WOW64; Trident/7.0; MDDCJS; rv:11.0) like Gecko")

		resp, e := client.Do(req)
		if e != nil {
			return e
		}

		// parse the page
		if doc, e = gq.NewDocumentFromResponse(resp); e != nil {
			return e
		}

		// regular expressions
		re_id := regexp.MustCompile("/torrent/(\\d+)-")
		re_category := regexp.MustCompile("gi-(\\w+)")
		re_category_cat3 := regexp.MustCompile("text-pink")
		re_adddate := regexp.MustCompile("(\\d+) (\\w+)")

		// Iterate on each record
		doc.Find("div.container-fluid div.row table.table.table-bordered tbody tr").Each(func(i int, s *gq.Selection) {
			//var se error

			link := s.Find("td > a.torrent-filename").First()
			if link.Length() == 0 {
				//p.logger.Println("ERROR: Link not found")
				return
			}

			href, hvalid := link.Attr("href")
			if !hvalid || href == "" {
				p.logger.Println("ERROR: Link not found")
				return
			}

			hu, se := url.Parse(href)
			if se != nil {
				p.logger.Printf("ERROR: %s", se)
				return
			}
			hu.Scheme = "http"
			hu.Host = "avistaz.to"

			// extract id from href
			idmatch := re_id.FindAllStringSubmatch(href, -1)
			if idmatch == nil || len(idmatch) < 1 || len(idmatch[0]) < 2 {
				p.logger.Printf("ID not found")
				return
			}

			lid := idmatch[0][1]

			category := s.Find("td > i.gi").First()
			if category.Length() == 0 {
				p.logger.Println("ERROR: Category not found")
				return
			}
			catclass, catvalid := category.Attr("class")
			if !catvalid || catclass == "" {
				p.logger.Println("ERROR: Cat class not found")
				return
			}

			catmatch := re_category.FindAllStringSubmatch(catclass, -1)
			if catmatch == nil || len(catmatch) < 1 || len(catmatch[0]) < 2 {
				p.logger.Printf("Category not found")
				return
			}

			catid := catmatch[0][1]
			if re_category_cat3.MatchString(catclass) {
				catid = "cat3"
			}

			seeder := s.Find("td").Eq(-3)
			if seeder.Length() == 0 {
				p.logger.Println("ERROR: Seeder not found")
				return
			}

			leecher := s.Find("td").Eq(-2)
			if leecher.Length() == 0 {
				p.logger.Println("ERROR: Leecher not found")
				return
			}
			complete := s.Find("td").Eq(-1)
			if complete.Length() == 0 {
				p.logger.Println("ERROR: Complete not found")
				return
			}

			comments := s.Find("td").Eq(-7)
			if comments.Length() == 0 {
				p.logger.Println("ERROR: Comments not found")
				return
			}

			adddate := s.Find("td").Eq(-6)
			if adddate.Length() == 0 {
				p.logger.Println("ERROR: Adddate not found")
				return
			}

			nseeder, se := strconv.ParseInt(seeder.Text(), 10, 32)
			if se != nil {
				p.logger.Printf("ERROR: %s", se)
				return
			}

			nleecher, se := strconv.ParseInt(leecher.Text(), 10, 32)
			if se != nil {
				p.logger.Printf("ERROR: %s", se)
				return
			}
			ncomplete := int64(0)
			if complete.Length() > 0 {
				ncomplete, se = strconv.ParseInt(complete.Text(), 10, 32)
				if se != nil {
					p.logger.Printf("ERROR: %s", se)
					return
				}
			}
			ncomments := int64(0)
			if comments.Length() > 0 && comments.Text() != "---" {
				ncomments, se = strconv.ParseInt(comments.Text(), 10, 32)
				if se != nil {
					p.logger.Printf("ERROR: %s", se)
					return
				}
			}

			//fmt.Printf("@@ %d - %d - %d - %d\n", nseeder, nleecher, ncomplete, ncomments)

			adddatematch := re_adddate.FindAllStringSubmatch(adddate.Text(), -1)
			if adddatematch == nil || len(adddatematch) < 1 || len(adddatematch[0]) < 3 {
				p.logger.Printf("Adddate not found")
				return
			}

			nduration, se := strconv.ParseInt(adddatematch[0][1], 10, 32)
			if se != nil {
				p.logger.Printf("ERROR parsing duration: %s", se)
				return
			}

			nadddate := time.Now()
			switch adddatematch[0][2] {
			case "minute", "minutes":
				nadddate = nadddate.Add(-1 * time.Minute * time.Duration(nduration))
			case "hour", "hours":
				nadddate = nadddate.Add(-1 * time.Hour * time.Duration(nduration))
			case "day", "days":
				nadddate = nadddate.Add(-1 * 24 * time.Hour * time.Duration(nduration))
			case "week", "weeks":
				nadddate = nadddate.Add(-1 * 7 * 24 * time.Hour * time.Duration(nduration))
			case "month", "months":
				nadddate = nadddate.Add(-1 * 30 * 24 * time.Hour * time.Duration(nduration))
			case "year", "years":
				nadddate = nadddate.Add(-1 * 365 * 24 * time.Hour * time.Duration(nduration))
			default:
				p.logger.Printf("ERROR determining duration: %s", adddatematch[0][2])
				return
			}

			//fmt.Println(nadddate.String())

			//fmt.Printf("%s: %s\n", strings.TrimSpace(link.Text()), lid)
			item, ok := p.List[lid]
			if !ok {
				item = fstoplib.NewItem()
				item.Id = lid
				item.Title = strings.TrimSpace(link.Text())
				item.Link = hu.String()
				item.Count = 0
				item.Category = catid
				item.AddDate = nadddate.Format("2006-01-02")
				item.Seeders = int32(nseeder)
				item.Leechers = int32(nleecher)
				item.Complete = int32(ncomplete)
				item.Comments = int32(ncomments)
				p.List[lid] = item
			}
			item.Count++
			posct++
			if sort == AVSORT_SEEDERS {
				item.SeedersPos = posct
			} else if sort == AVSORT_LEECHERS {
				item.LeechersPos = posct
			} else if sort == AVSORT_COMPLETE {
				item.CompletePos = posct
			}
		})
	}

	return nil
}
