package main

import (
	"log"
	"flag"
	"sync"
	"net/url"
)

var flag_url = flag.String("url", "", "Base URL")
var flag_workers = flag.Int("workers", 4, "Number of workers")


type Fetcher interface{
	FetchLinks(url string) ([]*url.URL, error)
}

type Crawler struct {
	id			int					// ID of the crawler for debugging.
	feedback	chan string			// Queue for pending Links.
	waitGroup	*sync.WaitGroup		// Manages active Links in Queue.
	fetcher		Fetcher				// Retrieves Links by URL.
}

func (c *Crawler) logf(fmt string, args ...interface{}) {
	log.Printf("%d: "+ fmt, append([]interface{}{c.id}, args...)...)
}

// Push the links to the feedback channel.
// Feeding the feedback channel may block, depending on the amount of
// new links and the capacity of the feedback channel.
func (c *Crawler) pushLinks(links []*url.URL) {
	for _, link := range links {
		c.feedback <- link.String()
	}
}

// Invoke Fetcher for Links in current URL and feed new links back.
func (c *Crawler) doCrawl(url string, feedback chan string) {
	c.logf("Crawling %s\n", url)

	links, err := c.fetcher.FetchLinks(url)

	if err != nil {
		c.logf("Error finding links in URL %s: %s\n", url, err)
		return
	}

	c.waitGroup.Add(len(links))
	go c.pushLinks(links)

	c.logf("Done searching in %s.\n", url)
}

func (c Crawler) Crawl() {
	crawl: for {
		select {
			// We've got work, do the work
			case toCrawl,ok := <-c.feedback:
				// The feedback channel was closed, there's no more
				// data to retrieve.
				if !ok {
					break crawl
				}

				c.doCrawl(toCrawl, c.feedback)

				c.waitGroup.Done()
		}
	}
}

func startCrawl(url string, workers int, fetcher Fetcher) {
	crawler := make([]*Crawler, workers)

	waitGroup := &sync.WaitGroup{}

	urlFeedback := make(chan string, workers)
	urlFeedback <- url
	waitGroup.Add(1)

	for i:=0; i < workers; i++ {
		crawler[i] = &Crawler{
			id:			i,
			feedback:	urlFeedback,
			waitGroup:	waitGroup,
			fetcher:	fetcher,
		}

		go crawler[i].Crawl()
	}

	waitGroup.Wait()

	log.Println("Crawling ended.")
}


func main() {
	flag.Parse()

	if *flag_url == "" {
		log.Fatal("No URL given.")
	}

	log.Println("Start crawl of", *flag_url)

	fetcher := NewHTTPFetcher(*flag_url)

	startCrawl(*flag_url, *flag_workers, fetcher)
}
