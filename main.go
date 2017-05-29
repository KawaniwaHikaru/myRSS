package main

import (
	"./parseEx"
	_ "github.com/go-sql-driver/mysql"
	"github.com/PuerkitoBio/goquery"
	"fmt"
	"time"
	"strings"
	"os"
	"log"
	"net/url"
	"path"
	"sync"
	"regexp"
	"database/sql"
	"github.com/go-sql-driver/mysql"
)

func parseMetalSucks() {
	type Song struct {
		Band  string
		Title string
	}

	songs := make([]Song, 0)

	p := parseEx.ParseEx{
		Url:    "http://metalsucks.net",
		Needle: ".sidebar-reviews article .content-block",
		Middleware: func(i int, s *goquery.Selection) {
			// For each item found, get the band and title
			song := Song{
				Band:  s.Find("a").Text(),
				Title: s.Find("i").Text(),
			}
			songs = append(songs, song)
		},
	}

	p.Scan()
	fmt.Println(songs)
}

type Article struct {
	Url     string
	Title   string
	Content string
	Created time.Time
}

func ArticleDAO(pageCh chan Article) {

	defer db.Close()

	var err error
	db, err = sql.Open("mysql", "myrss@/myrss?charset=utf8")
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := db.Prepare("INSERT articles SET url=?,title=?,content=?,created=?")
	if err != nil {
		log.Fatalln(err)
	}

	var isDone bool
	for !isDone {
		select {
		case page := <-pageCh:
			if _, err := stmt.Exec(page.Url, page.Title, page.Content, page.Created); err != nil {
				if mysqlError, ok := err.(*mysql.MySQLError); ok {
					if mysqlError.Number == 1062 {
						//duplicate unique
						fmt.Println(`Duplicate Record`, page.Url)
					}
				}
			} else {
				fmt.Println(page.Title, page.Created)
			}
		}
	}
}

func parseShinto(targetURL url.URL) {

	p := parseEx.ParseEx{
		Url:    targetURL.String(),
		Needle: "div#inner_content_ver2",
		Middleware: func(i int, s *goquery.Selection) {
			// parse the Time
			timeStr := s.Find(".title_noline").Next().Text()
			var publishedTime time.Time
			var err error

			switch len(timeStr) {
			case 12:
				publishedTime, err = time.Parse("[2006-01-02]", timeStr)
			case 17:
				publishedTime, err = time.Parse("[2006-01-02 15:04]", timeStr)
			default:
				publishedTime = time.Now()
			}
			if err != nil {
				fmt.Println(err)
			}

			content, _ := s.Find(".content").Html()

			// For each item found, get the band and title
			page := Article{
				Url:     targetURL.String(),
				Title:   strings.TrimSpace(s.Find(".title_noline").Text()),
				Content: strings.TrimSpace(content),
				Created: publishedTime,
			}

			// pass the new page to the article Channel
			articleCh <- page
		},
	}

	p.Scan()
}

// matching for specific path
// eg. /1165158/2017-05-24/post-土改三年 逾百萬地主亡/
var r = regexp.MustCompile(`/(\d*)/(\d{4}-\d{2}-\d{2})/(.*)`)
var db *sql.DB

func parsePage(srcURL string) (links map[string]url.URL) {

	// make a hash
	links = make(map[string]url.URL)

	srcPageURL, err := url.Parse(srcURL)
	if err != nil {
		log.Fatalln("Can not parse URL")
	}

	// make sure we have schema
	if srcPageURL.Scheme == "" {
		srcPageURL.Scheme = "http"
	}

	// get the Document
	resp, err := goquery.NewDocument(srcPageURL.String())
	if err != nil {
		log.Fatalln(`ERROR: Failed to parse Document `, srcPageURL)
	}

	// use CSS selector found with the browser inspector
	// for each, use index and item
	resp.Find("a").Each(func(index int, item *goquery.Selection) {
		link, _ := item.Attr("href")
		//linkTarget, _ := item.Attr("target")

		// see if we can parse this href
		href, err := url.Parse(link)
		if err != nil {
			// can't parse this link, move to next
			return
		}
		// fill in the schema with targetURL
		if href.Scheme == "" {
			href.Scheme = srcPageURL.Scheme
		}

		//filter out empty link or none http links
		if href.Scheme != "http" || href.Path == "" || href.Path == "#" {
			return
		}

		if !r.MatchString(href.Path) {
			return
		}

		// get the first //
		fields := strings.Split(href.Path, "/")
		href.Path = fields[1] + "/"
		//newURL, err := url.Parse(href.Scheme + "://" + href.Host + "/" + fields[1] + "/")
		//fmt.Println(newURL)

		links[href.String()] = *href
	})

	return
}

//var doneCh = make(chan bool)
var articleCh = make(chan Article)

func init() {

	// spawn the dao
	go ArticleDAO(articleCh)

}

func main() {
	var wg sync.WaitGroup

	if len(os.Args) < 2 {
		log.Fatalln("ERROR : Less Args\nCommand should be of type : " + path.Base(os.Args[0]) + " [folder to save] [websites]\n\n")
	}

	// use a hash-map to hold the found links
	links := parsePage(os.Args[1])

	//fmt.Println(len(links))
	wg.Add(len(links))
	for _, value := range links {

		// pass the value into the go func,
		// do not access variable value inside the anonymous function
		// it is not always pointing to the same object
		go func(u url.URL) {
			parseShinto(u)
			wg.Done()
		}(value)
	}

	wg.Wait()

	// tell DAO to exit
	//doneCh <- true
}
