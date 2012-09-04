package main

import (
	"testing"
	"net/url"
)

type FakeFetcher struct{
	linkMap		map[string][]string
	visited		map[string]bool
}

func newFakeFetcher() *FakeFetcher {
	return &FakeFetcher{
		make(map[string][]string),
		make(map[string]bool),
	}
}

func (f *FakeFetcher) AddLinks(url string, followups []string) {
	f.linkMap[url] = followups
	f.visited[url] = false

	for _, link := range followups {
		f.visited[link] = false
	}
}

func (f *FakeFetcher) Summary(t *testing.T) {
	for url,visited := range f.visited {
		if !visited {
			t.Logf("URL %s NOT visited.\n", url)
		} else {
			t.Logf("URL %s visited.\n", url)
		}
	}
}

func mapToURL(sURLS []string) []*url.URL {
	urls := make([]*url.URL, len(sURLS))

	for i, surl := range sURLS {
		urls[i], _ = url.Parse(surl)
	}

	return urls
}

func (f FakeFetcher) FetchLinks(curl string) ([]*url.URL, error) {
	if f.visited[curl] {
		return nil, nil
	}

	f.visited[curl] = true

	return mapToURL(f.linkMap[curl]), nil
}


func TestSimpleCrawl(t *testing.T) {
	ff := newFakeFetcher()

	ff.AddLinks("http://localhost/", []string{
		"http://localhost/sub1/",
		"http://localhost/sub2/index.php",
	})

	ff.AddLinks("http://localhost/sub1/", []string{
		"http://localhost/sub1/",
		"http://localhost/sub2/image.jpg",
	})

	ff.AddLinks("http://localhost/sub2/index.php", []string{
		"http://localhost/sub3/",
		"http://someotherhost/foo.html",
	})

	startCrawl("http://localhost/", 1, ff)

	ff.Summary(t)
}

