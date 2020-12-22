package epinfo

import (

	//for sending HTTP requests

	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html" //for URL formatting
)

// EpisodeInfo struct holds info about a episode
type EpisodeInfo struct {
	Label  string
	Series string
	URL    string
}

func bodyNode(url string) *html.Node {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	body, err := html.Parse(strings.NewReader(string(bodyBytes)))
	if err != nil {
		log.Fatal(err)
	}
	return body
}

// SeriesEpisodes return all episodes found on a given url
func SeriesEpisodes(pageURL string) []EpisodeInfo {
	pageVisited := make(map[string]bool)
	pageVisited[pageURL] = false
	body := bodyNode(pageURL)
	episodes := []EpisodeInfo{}
	for url, visited := range pageVisited {
		if !visited {
			var f func(*html.Node)
			// Depth-first order processing
			f = func(node *html.Node) {
				if node.Type == html.ElementNode && node.Data == "a" {
					href := ""
					label := ""
					series := "none"
					for _, attr := range node.Attr {
						switch attr.Key {
						case "href":
							if strings.Contains(attr.Val, "/iplayer/episode/") {
								href = "https://www.bbc.co.uk" + attr.Val
							} else if strings.Contains(attr.Val, "?page=") {
								_, ok := pageVisited[url+attr.Val]
								if !ok {
									pageVisited[url+attr.Val] = false
								}
							}
						case "aria-label":
							label = attr.Val
						case "data-bbc-container":
							series = attr.Val
						}
					}
					if href != "" && label != "" && series != "contextual-cta" {
						newEpisode := EpisodeInfo{label, series, href}
						episodes = append(episodes, newEpisode)
					}
				}
				for c := node.FirstChild; c != nil; c = c.NextSibling {
					f(c)
				}
			}
			f(body)
		}
	}
	return episodes
}

// SeriesURLs returns all links to series web pages
func SeriesURLs(pageURL string) map[string]string {
	body := bodyNode(pageURL)
	series := make(map[string]string)
	var f func(*html.Node)
	// Depth-first order processing
	f = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "a" {
			href := ""
			seriesName := ""
			for _, attr := range node.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "series-nav__button") {
					seriesName = (node.FirstChild).FirstChild.Data
				} else if attr.Key == "href" && strings.Contains(attr.Val, "?seriesId=") {
					href = "https://www.bbc.co.uk" + attr.Val
				}
			}
			if href != "" && seriesName != "" {
				existingURL, ok := series[seriesName]
				if !ok {
					series[seriesName] = href
				} else if existingURL != href {
					log.Fatalf("Series Name: %s has already link: %s but also found: %s", seriesName, existingURL, href)
				}
			}
		} else if node.Type == html.ElementNode && node.Data == "span" {
			for _, attr := range node.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "series-nav__button") {
					series[(node.FirstChild).FirstChild.Data] = pageURL
				}
				break
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(body)
	return series
}

// AllEpisodesInfo returns an array of all episodes of a given BBC iPlayer URL.
// Sorted from first to last.
func AllEpisodesInfo(pageURL string) map[string][]EpisodeInfo {
	urlSuffixes := []string{"?page=", "?seriesId="}
	for _, s := range urlSuffixes {
		suffixIndex := strings.LastIndex(pageURL, s)
		if suffixIndex != -1 {
			pageURL = pageURL[:suffixIndex]
		}
	}
	foundSeriesURLs := SeriesURLs(pageURL)
	allSeriesEpisodes := make(map[string][]EpisodeInfo)
	if len(foundSeriesURLs) == 0 {
		foundSeriesURLs["none"] = pageURL
	}
	for seriesName, sURL := range foundSeriesURLs {
		allSeriesEpisodes[seriesName] = SeriesEpisodes(sURL)
	}
	defaultSeriesEP, ok := allSeriesEpisodes["none"]
	if ok && defaultSeriesEP[0].Series != "none" {
		ep := defaultSeriesEP[0]
		allSeriesEpisodes[ep.Series] = defaultSeriesEP
		delete(allSeriesEpisodes, "none")
	}
	return allSeriesEpisodes
}
