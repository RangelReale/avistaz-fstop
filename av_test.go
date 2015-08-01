package avistaz

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"
)

func TestDownload(t *testing.T) {
	return

	// download the page
	u, e := url.Parse(fmt.Sprintf("https://avistaz.to/torrents?&order=%s&sort=%s&page=%d", AVSORT_SEEDERS, AVSORTBY_DESCENDING, 1))
	if e != nil {
		t.Error(e)
		return
	}

	cookies, _ := cookiejar.New(nil)

	cookies.SetCookies(u, []*http.Cookie{
		&http.Cookie{Name: "avistaz_session", Value: "", Path: "/", Domain: "avistaz.to"},
	})

	client := &http.Client{
		Jar: cookies,
	}

	req, e := http.NewRequest("GET", u.String(), nil)
	if e != nil {
		t.Error(e)
		return
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.3; WOW64; Trident/7.0; MDDCJS; rv:11.0) like Gecko")

	resp, e := client.Do(req)
	if e != nil {
		t.Error(e)
		return
	}

	defer resp.Body.Close()

	contents, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		t.Error(e)
		return
	}

	fmt.Println(string(contents))
}

func TestFetcher(t *testing.T) {
	c := NewConfig()
	c.Avistaz_session = ""
	f := NewFetcher(c)
	f.SetLogger(log.New(ioutil.Discard, "", 0))
	i, err := f.Fetch()
	if err != nil {
		t.Error(err)
		return
	}

	if len(i) == 0 {
		t.Error("No data returned from parser")
	}
}
