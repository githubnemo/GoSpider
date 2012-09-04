# GoSpider

This is a really simple web spider.

The purpose of this program is to fetch all local links on a
website and follow them recursively. It's just a visitor and
nothing else.

To speed things up, this whole thing works concurrently.

## Usage

	$ go build
	$ ./crawl -workers=16 -url=http://localhost/

