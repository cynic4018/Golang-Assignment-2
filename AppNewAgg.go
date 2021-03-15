package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

var wg sync.WaitGroup

type SitemapIndex struct {
	Locations []string `xml:"sitemap>loc"`
}

type News struct {
	Titles    []string `xml:"url>news>title"`
	Keywords  []string `xml:"url>news>keywords"`
	Locations []string `xml:"url>loc"`
}

type NewsMap struct {
	Keyword  string
	Location string
}

type NewsAggPage struct {
	Title string
	News  map[string]NewsMap
}

func indexHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "<h1>Whoa, GO is nest!</h1>")
}

func newRoutine(c chan News, location string) {

	defer wg.Done()

	var n News

	location = strings.TrimSpace(location)
	res, _ := http.Get(location)
	bytes, _ := ioutil.ReadAll(res.Body)
	xml.Unmarshal(bytes, &n)
	res.Body.Close()

	c <- n
}

func newAggHandler(w http.ResponseWriter, r *http.Request) {

	var s SitemapIndex
	news_map := make(map[string]NewsMap)

	res, _ := http.Get("https://www.washingtonpost.com/news-sitemaps/index.xml")
	bytes, _ := ioutil.ReadAll(res.Body)
	xml.Unmarshal(bytes, &s)
	res.Body.Close()

	queue := make(chan News, 30)

	for _, Location := range s.Locations {
		wg.Add(1)
		go newRoutine(queue, Location)
	}

	wg.Wait()
	close(queue)

	for elem := range queue {
		for idx, _ := range elem.Keywords {
			news_map[elem.Titles[idx]] = NewsMap{elem.Keywords[idx], elem.Locations[idx]}
		}

	}

	p := NewsAggPage{Title: "Amazing News Aggregator", News: news_map}
	t, _ := template.ParseFiles("basictemplating.html")
	t.Execute(w, p)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/agg/", newAggHandler)
	http.ListenAndServe(":8000", nil)
}
