package cli

import (
	"fmt"
	"log"
	"strings"

	"github.com/gandalf15/iplayer/epinfo"
)

// Cli runs command line interface
func Cli(url *string, audioDescribed *bool, signLang *bool) {

	if *url == "" {
		log.Fatal("usage: ./iplayer -url=[iPlayer URL with episodes]")
	} else {
		allSeries := epinfo.AllEpisodesInfo(*url, *audioDescribed, *signLang)
		var epLinks []string
		for _, v := range allSeries {
			for _, epi := range v {
				epLinks = append(epLinks, epi.URL)
			}
		}
		fmt.Print(strings.Join(epLinks, "\n"))
	}
}
