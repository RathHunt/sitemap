package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

type Link struct {
	XMLName xml.Name `xml:"url"`
	Addr    string   `xml:"loc"`
}

type Sitemap struct {
	XMLName xml.Name `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
	Links   []Link
}

func getLinks(n *html.Node) []string {

	var links []string

	if n.Type == html.ElementNode && n.Data == "a" {
		for _, a := range n.Attr {
			if a.Key == "href" {
				links = append(links, a.Val)
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		links = append(links, getLinks(c)...)
	}

	return links
}

func baseDomain(url string) string {
	if strings.HasPrefix(url, "/") {
		return "/"
	}

	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}

	re := *regexp.MustCompile(`https?:\/\/www.+?\/`)
	s := re.FindStringSubmatch(url)

	if s == nil {
		return "/"
	}

	return s[0]
}

func Crawl(url string, visited *map[string]bool, currentDepth int, maxDepth *int) ([]Link, error) {
	if currentDepth == *maxDepth {
		return nil, nil
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	rawlinks := getLinks(doc)

	var links []Link

	urlBaseDomain := baseDomain(url)

	for _, link := range rawlinks {
		if dom := baseDomain(link); dom == urlBaseDomain || dom == "/" && !(*visited)[link] {
			if dom == "/" {
				link = urlBaseDomain + link
			}
			links = append(links, Link{Addr: link})
			(*visited)[link] = true
			moar, err := Crawl(link, visited, currentDepth+1, maxDepth)
			if err != nil {
				return nil, err
			}
			links = append(links, moar...)
		}
	}
	return links, err
}

func init() {

}

func main() {
	url := flag.String("url", "https://www.example.com", "the site to crawl")
	depth := flag.Int("depth", 0, "max page depth to crawl (0 to infinite)")
	flag.Parse()

	visited := map[string]bool{}

	links, err := Crawl(*url, &visited, 1, depth)
	if err != nil {
		panic(err)
	}

	sitemap := Sitemap{Links: links}
	data, _ := xml.MarshalIndent(sitemap, "", " ")
	data = []byte(xml.Header + string(data))
	fmt.Printf("%s\n", data)
}
