package main

import (
	"log"
	"strings"
	"net/http"
	"net/url"
	"io/ioutil"
	"regexp"
)

const anchorRegexp = `<[aA][^>]*>`
const baseRegexp = `<[bB][aA][sS][eE][^>]*>`
const hrefRegexp = `(?i:href)=(?:'([^']+)'|"([^"]+)")`

var anchorPattern = regexp.MustCompile(anchorRegexp)
var basePattern = regexp.MustCompile(baseRegexp)
var hrefPattern = regexp.MustCompile(hrefRegexp)

func readUrl(url string) (string,error) {
	res, err := http.Get(url)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	return string(body), nil
}

func stripFileFromURL(furl string) string {
	parsedUrl, err := url.Parse(furl)

	if err != nil {
		return furl
	}

	paths := strings.Split(parsedUrl.Path, "/")
	parsedUrl.Path = strings.Join(paths[0:len(paths)-1], "/")+"/"
	parsedUrl.RawQuery = ""
	parsedUrl.Fragment = ""

	return parsedUrl.String()
}

func findBaseURL(currentURL string, content string) string {
	allBaseTags := basePattern.FindAllString(content, -1)

	baseURL := currentURL

	if len(allBaseTags) > 0 {
		href := hrefPattern.FindStringSubmatch(allBaseTags[len(allBaseTags)-1])

		if len(href) == 3 {
			baseURL = href[2]
		}
	}
	// Strip the last element of the path. Things can
	// only be relative to directories, not files.
	return stripFileFromURL(baseURL)
}

func findLinks(content string) ([]*url.URL,error) {
	allAnchors := anchorPattern.FindAllString(content, -1)

	links := make([]*url.URL, 0, len(allAnchors))

	for _, anchor := range allAnchors {
		href := hrefPattern.FindStringSubmatch(anchor)

		if len(href) != 3 {
			continue
		}

		parsed, err := url.Parse(href[2])

		if err != nil {
			log.Printf("Error while parsing '%s': %s.\n", href[2], err)
			continue
		}

		links = append(links, parsed)
	}

	return links, nil
}

func applyBaseURL(baseURL string, links []*url.URL) []*url.URL {
	parsedBaseURL,err := url.Parse(baseURL)

	if err != nil {
		log.Printf("HTTP Fetcher: Can't parse baseURL '%s': %s.\n",
			baseURL, err)
		return nil
	}

	for _, link := range links {
		if link.Host != "" || len(link.Path) == 0 {
			continue
		}

		// The relative to host is implicitly handled
		// by setting the host/scheme in every case.
		link.Host = parsedBaseURL.Host
		link.Scheme = parsedBaseURL.Scheme

		// Relative to Base URL's Path is handled here.
		if link.Path[0] != '/' {
			link.Path = parsedBaseURL.Path + link.Path
		}
	}

	return links
}



type HTTPFetcher struct{
	startURL		*url.URL
	visited			map[string]struct{}
}

func (h *HTTPFetcher) filterExternalsAndVisited(links []*url.URL) []*url.URL {
	goods := make([]*url.URL, 0, len(links))

	for _, link := range links {
		if link.Host != h.startURL.Host {
			continue
		}

		if _, ok := h.visited[link.String()]; ok {
			continue
		}

		goods = append(goods, link)
	}

	return goods
}

func (h *HTTPFetcher) markVisited(links []*url.URL) {
	for _, link := range links {
		h.visited[link.String()] = struct{}{}
	}
}

func (h *HTTPFetcher) FetchLinks(url string) ([]*url.URL, error) {
	content, err := readUrl(url)

	if err != nil {
		return nil, err
	}

	links, err := findLinks(content)

	if err != nil {
		return nil, err
	}

	baseURL := findBaseURL(url, content)

	links = applyBaseURL(baseURL, links)

	links = h.filterExternalsAndVisited(links)

	h.markVisited(links)

	return links, nil
}

func NewHTTPFetcher(startURL string) *HTTPFetcher {
	parsed, err := url.Parse(startURL)

	if err != nil {
		log.Fatal("Can't parse URL '%s': %s\n", startURL, err)
	}

	return &HTTPFetcher{
		parsed,
		make(map[string]struct{}),
	}
}
