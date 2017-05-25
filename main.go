package main

import (
    "./parseEx"
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
)

func parseMetalSucks() {
    type Song struct {
        Band  string
        Title string
    }

    songs := make([]Song, 0)

    p := parseEx.ParseEx{
        Url: "http://metalsucks.net",
        Needle: ".sidebar-reviews article .content-block",
        Middleware: func(i int, s *goquery.Selection) {
            // For each item found, get the band and title
            song := Song{
                Band: s.Find("a").Text(),
                Title : s.Find("i").Text(),
            }
            songs = append(songs, song)
        },
    }

    p.Scan()
    fmt.Println(songs)
}

func parseShinto(urlObj url.URL) {
    type NewsContent struct {
        Url     string
        Title   string
        Content string
        Created time.Time
    }

    p := parseEx.ParseEx{
        Url: urlObj.String(),
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
            page := NewsContent{
                Url: urlObj.String(),
                Title: strings.TrimSpace(s.Find(".title_noline").Text()),
                Content: strings.TrimSpace(content),
                Created: publishedTime,
            }
            fmt.Print(page)
        },
    }

    p.Scan()
}

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

        // matching for specific path
        // eg. /1165158/2017-05-24/post-土改三年 逾百萬地主亡/
        r := regexp.MustCompile(`/(\d*)/(\d{4}-\d{2}-\d{2})/(.*)`)
        if !r.MatchString(href.Path) {
            return
        }

        // all passed, now push it into the hashmap
        links[href.Path] = *href
    })

    return
}

func main() {

    var wg sync.WaitGroup

    if len(os.Args) < 2 {
        log.Fatalln("ERROR : Less Args\nCommand should be of type : " + path.Base(os.Args[0]) + " [folder to save] [websites]\n\n")
    }

    seedUrls := os.Args[1:]

    // use a hash-map to hold the found links
    links := parsePage(seedUrls[0])

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
}
