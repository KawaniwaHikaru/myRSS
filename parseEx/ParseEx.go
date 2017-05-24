package parseEx

import (
    "github.com/PuerkitoBio/goquery"
    "log"
)

type ParseEx struct {
    Url        string
    Needle     string
    Middleware func(i int, s *goquery.Selection)
}

func (p *ParseEx)Scan() {

    doc, err := goquery.NewDocument(p.Url)
    if err != nil {
        log.Fatal(err)
    }

    // Find the review items
    doc.Find(p.Needle).Each(p.Middleware)
}

