package main

import (
    "github.com/PuerkitoBio/goquery"
    "log"
    "fmt"
    "time"
    "strings"
)

type ParseEx struct {
    url        string
    needle     string
    middleware func(i int, s *goquery.Selection)
}

func (p *ParseEx)Scan() {

    doc, err := goquery.NewDocument(p.url)
    if err != nil {
        log.Fatal(err)
    }

    // Find the review items
    doc.Find(p.needle).Each(p.middleware)
}

func parseMetalSucks() {
    type Song struct {
        Band  string
        Title string
    }

    songs := make([]Song, 0)

    p := ParseEx{
        url: "http://metalsucks.net",
        needle: ".sidebar-reviews article .content-block",
        middleware: func(i int, s *goquery.Selection) {
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

    type Page struct {
        Title   string
        Content string
        Created time.Time
    }

    p := ParseEx{
        url: "http://vancouver.singtao.ca/1164488/2017-05-23/post-%e5%86%b7%e5%b3%b0%e7%a7%bb%e5%90%91%e5%8c%97%e5%b2%b8-%e4%bd%8e%e9%99%b8%e5%b9%b3%e5%8e%9f%e5%90%b9%e5%bc%b7%e9%a2%a8/?variant=zh-hk",
        needle: "div#inner_content_ver2",
        middleware: func(i int, s *goquery.Selection) {

            timeStr := s.Find(".title_noline").Next().Text()
            fmt.Println(timeStr)
            ctime, err := time.Parse("[2006-1-2 15:4]", timeStr)
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
