package epinfo

import (

	//for sending HTTP requests

	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html" //for URL formatting
)

// EpisodeInfo struct holds info about an episode
type EpisodeInfo struct {
	TvShow                   *string
	Label, Series, URL       string
	AudioDescribed, SignLang bool
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

// SeriesEpisodes return all episodes found on a given url.
// You can select if you want to include audio described and sign language links.
func SeriesEpisodes(pageURL string, audioDescribed bool, signLang bool, ch chan []EpisodeInfo) {
	pageVisited := make(map[string]bool)
	pageVisited[pageURL] = true
	body := bodyNode(pageURL)
	episodes := []EpisodeInfo{}
	tvShow := ""
	tvShowFound := false
	var f func(*html.Node)
	// Depth-first order processing
	f = func(node *html.Node) {
		if node.Type == html.ElementNode {
			if node.Data == "a" {
				href := ""
				label := ""
				series := "none"
				for _, attr := range node.Attr {
					switch attr.Key {
					case "href":
						if strings.Contains(attr.Val, "/iplayer/episode/") {
							href = "https://www.bbc.co.uk" + attr.Val
						} else if strings.Contains(attr.Val, "?page=") {
							i := strings.LastIndex(attr.Val, "?page=")
							_, ok := pageVisited[pageURL+attr.Val[i:]]
							if !ok {
								pageVisited[pageURL+attr.Val[i:]] = false
							}
						}
					case "aria-label":
						label = attr.Val
					case "data-bbc-container":
						series = attr.Val
					}
				}
				if href != "" && label != "" && series != "contextual-cta" {
					if strings.Contains(href, "/ad/") {
						if audioDescribed {
							newEpisode := EpisodeInfo{&tvShow, label, series, href, true, false}
							episodes = append(episodes, newEpisode)
						}
					} else if strings.Contains(href, "/sign/") {
						if signLang {
							newEpisode := EpisodeInfo{&tvShow, label, series, href, false, true}
							episodes = append(episodes, newEpisode)
						}
					} else {
						newEpisode := EpisodeInfo{&tvShow, label, series, href, false, false}
						episodes = append(episodes, newEpisode)
					}

				}
			} else if node.Data == "h1" {
				for _, attr := range node.Attr {
					if !tvShowFound {
						if attr.Key == "class" && strings.Contains(attr.Val, "title") {
							tvShow = node.FirstChild.Data
						}
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(body)

	for url, visited := range pageVisited {
		if !visited {
			body := bodyNode(url)
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
								i := strings.LastIndex(attr.Val, "?page=")
								_, ok := pageVisited[url+attr.Val[i:]]
								if !ok {
									pageVisited[url+attr.Val[i:]] = false
								}
							}
						case "aria-label":
							label = attr.Val
						case "data-bbc-container":
							series = attr.Val
						}
					}
					if href != "" && label != "" && series != "contextual-cta" {
						if strings.Contains(href, "/ad/") {
							if audioDescribed {
								newEpisode := EpisodeInfo{&tvShow, label, series, href, true, false}
								episodes = append(episodes, newEpisode)
							}
						} else if strings.Contains(href, "/sign/") {
							if signLang {
								newEpisode := EpisodeInfo{&tvShow, label, series, href, false, true}
								episodes = append(episodes, newEpisode)
							}
						} else {
							newEpisode := EpisodeInfo{&tvShow, label, series, href, false, false}
							episodes = append(episodes, newEpisode)
						}
					}
				}
				for c := node.FirstChild; c != nil; c = c.NextSibling {
					f(c)
				}
			}
			f(body)
		}
	}
	ch <- episodes
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

// AllEpisodesInfo returns a map of all series, if exist, and their episodes of a given BBC iPlayer URL.
// There is no guarantee that the episodes are sorted. Fetched from top to bottom of the page.
// It depends on the BBC iPlayer web page how the episodes are presented.
// signLang set true if you want to include sign language links.
// audioDescribed set true if you want to include audio descriabed links.
func AllEpisodesInfo(pageURL string, audioDescribed bool, signLang bool) map[string][]EpisodeInfo {
	urlSuffixes := []string{"?page=", "?seriesId="}
	for _, s := range urlSuffixes {
		suffixIndex := strings.LastIndex(pageURL, s)
		if suffixIndex != -1 {
			pageURL = pageURL[:suffixIndex]
		}
	}
	foundSeriesURLs := SeriesURLs(pageURL)
	if len(foundSeriesURLs) == 0 {
		foundSeriesURLs["none"] = pageURL
	}
	ch := make(chan []EpisodeInfo, len(foundSeriesURLs))
	for _, sURL := range foundSeriesURLs {
		go SeriesEpisodes(sURL, audioDescribed, signLang, ch)
	}
	allSeriesEpisodes := make(map[string][]EpisodeInfo)
	for range foundSeriesURLs {
		epArr := <-ch
		if len(epArr) > 0 {
			allSeriesEpisodes[epArr[0].Series] = epArr
		}
	}
	return allSeriesEpisodes
}
