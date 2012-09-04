package main

import (
	"testing"
	"net/url"
)

const exampleHTML = `
<html>
	<head>
		<title>Foo</title>
	</head>
	<body>
	<a href="/fullhost/relative">Full Host Relative</a>
	<a href="url/relative/">URL Relative</a>
	<a href="http://simplelink" name="slink">Slink</a>
	<a href="http://dis/is/a/link" name="foo">Bar</a>
	<b><A class="foo" HREF="http://sumlinkyo/foo"/></b>
	</body>
</html>`

const baseURLexample = `
<body>
	<base href="http://localhost/"/>
	<a href="/host/url/relative">foo</a>
	<a href="base/url/relative">bar</a>
	<a href="http://localhost/external/">baz</a>
</body>`


func TestReadURL(t *testing.T) {
	content, err := readUrl("http://localhost/")

	if err != nil {
		t.Fatal(err)
	}

	t.Log(content)
}

func findLink(t *testing.T, links []*url.URL, link string) {
	for _, ref := range links {
		//t.Logf("compare '%s' with '%s'\n", ref, link)
		if ref.String() == link {
			return
		}
	}
	t.Fatalf("Link '%s' not found in %v.\n", link, links)
}

func TestFindLinks(t *testing.T) {
	content := exampleHTML
	links, err := findLinks(content)

	if err != nil {
		t.Fatal(err)
	}

	findLink(t, links, "http://simplelink")
	findLink(t, links, "http://dis/is/a/link")
	findLink(t, links, "http://sumlinkyo/foo")
}

func TestFindBaseURL(t *testing.T) {
	content := baseURLexample

	// base tag present, result should be the url in the tag
	baseURL := findBaseURL("http://localhost/current/", content)

	if baseURL != "http://localhost/" {
		t.Fatalf("baseURL '%s' not matching with needle.\n", baseURL)
	}

	t.Log("BaseURL:", baseURL)

	// No base tag present, take the current url as base
	content = exampleHTML

	baseURL = findBaseURL("http://localhost/current/", content)

	if baseURL != "http://localhost/current/" {
		t.Fatalf("baseURL '%s' not matching iwth needle.\n", baseURL)
	}

	t.Log("BaseURL:", baseURL)
}

func TestApplyBaseURL(t *testing.T) {
	links := []string{
		"http://localhost/external/",
		"/host/relative",
		"relative",
	}

	applied := applyBaseURL("http://localhost/baseURL/", mapToURL(links))

	findLink(t, applied, links[0])
	findLink(t, applied, "http://localhost/host/relative")
	findLink(t, applied, "http://localhost/baseURL/relative")
}

func TestReadAndFetch(t *testing.T) {
	content, err := readUrl("http://localhost/")

	if err != nil {
		t.Fatal(err)
	}

	links, err := findLinks(content)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(links)
}

func TestStripFileFromURL(t *testing.T) {
	canditates := [][]string{
		[]string{"http://localhost/", "http://localhost/"},
		[]string{"http://localhost/foo", "http://localhost/"},
		[]string{"http://localhost/foo/", "http://localhost/foo/"},
		[]string{"http://localhost/foo/?test", "http://localhost/foo/"},
	}

	for _, pair := range canditates {
		res := stripFileFromURL(pair[0])

		if res != pair[1] {
			t.Fatalf("Candidate '%s': Got '%s', expected '%s'.\n",
				pair[0], res, pair[1])
		}
	}
}
