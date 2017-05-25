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

func parseShinto(strUrl string) {
    type NewsContent struct {
        Url     string
        Title   string
        Content string
        Created time.Time
    }

    p := parseEx.ParseEx{
        Url: strUrl,
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
                Url: strUrl,
                Title: strings.TrimSpace(s.Find(".title_noline").Text()),
                Content: strings.TrimSpace(content),
                Created: publishedTime,
            }
            fmt.Print(page)
        },
    }

    p.Scan()
}

func Crawl(strUrl string, srcCh chan string, chFinished chan bool) {

    fmt.Println("Fetching Page..")
    defer func() {
        fmt.Println("Done Crawling...")
        chFinished <- true
    }()

    // main document
    mainDoc, err := url.Parse(strUrl)
    if (err != nil) {
        return
    }

    // get the Document
    resp, err := goquery.NewDocument(strUrl)
    if err != nil {
        log.Fatalln(`ERROR: Failed to crawl `, strUrl)
    }


    // use CSS selector found with the browser inspector
    // for each, use index and item
    resp.Find("a").Each(func(index int, item *goquery.Selection) {
        //link, _ := item.Find("img").Attr("src")
        link, _ := item.Attr("href")
        target, _ := item.Attr("target")
        href, err := url.Parse(link)

        if err != nil || link == "" || link == "#" || target != "_blank" {
            return
        }

        if !strings.HasPrefix(link, "http") {
            link = mainDoc.Scheme + ":" + link
        }

        if href.Host == mainDoc.Host {
            srcCh <- link
        }

    })
}


/**
 Unnecessary complex for parsing one page.
 might as well support multiple init URL
 */
func parsePage(strUrl string) (links map[string]int) {

    // Channels
    chLinks := make(chan string)
    chFinished := make(chan bool)

    // check if we start with http
    if !strings.HasPrefix(strUrl, "http") {
        strUrl = "http://" + strUrl
    }

    // probably want to rewrite this
    // we are only parsing one page
    // should do the hash link collection inside
    // the Crawl
    go Crawl(strUrl, chLinks, chFinished)

    var isFinished bool

    for !isFinished {
        select {
        case url := <-chLinks:
            links[url]++

        case isFinished = <-chFinished:

        }
    }

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

    // spawn Go Routines for all hash maps
    // should do some kind of throttling
    wg.Add(len(links))
    for key := range links {
        // run the parser and call wg.Done
        // use Anonymous function to avoid passing the
        // wg into the function
        go func() {
            //fmt.Println(key)
            parseShinto(key)
            wg.Done()
        }()
    }

    wg.Wait()
}
