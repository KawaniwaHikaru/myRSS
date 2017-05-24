package main

import (
    "./parseEx"
    "github.com/PuerkitoBio/goquery"
    "fmt"
    "time"
    "strings"
    "net/url"
)

func parseMetalSucks() {
    type Song struct {
        Band  string
        Title string
    }

    songs := make([]Song, 0)

    p := ParseEx.ParseEx{
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

func main() {
}

func parseShinto() {
    type Page struct {
        Url     url.URL
        Title   string
        Content string
        Created time.Time
    }

    p := parseEx.ParseEx{
        Url: "http://vancouver.singtao.ca/1165366/2017-05-24/post-%E4%B8%80%E4%BB%A3%E4%BF%A0%E5%A5%B3%E4%BA%8E%E7%B4%A0%E7%A7%8B%E9%95%B7%E7%9C%A0%E4%B8%89%E8%97%A9%E5%B8%82/?variant=zh-hk",
        Needle: "div#inner_content_ver2",
        Middleware: func(i int, s *goquery.Selection) {
            // parse the Time
            timeStr := s.Find(".title_noline").Next().Text()
            var ctime time.Time
            var err error

            switch len(timeStr) {
            case 12:
                ctime, err = time.Parse("[2006-01-02]", timeStr)
            case 17:
                ctime, err = time.Parse("[2006-01-02 15:04]", timeStr)
            default:
                ctime = time.Now()
            }
            if err != nil {
                fmt.Println(err)
            }

            // For each item found, get the band and title
            page := Page{
                Title: strings.TrimSpace(s.Find(".title_noline").Text()),
                Content: strings.TrimSpace(s.Find(".content").Text()),
                Created: ctime,
            }
            fmt.Print(page)
        },
    }

    p.Scan()
}
