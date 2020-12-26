package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/gandalf15/iplayer/epinfo"
)

func main() {
	urlPtr := flag.String("url", "", "-url=[iPlayer URL with episodes]")
	flag.Parse()
	if *urlPtr == "" {
		log.Fatal("usage: ./iplayer -url=[iPlayer URL with episodes]")
	}
	allSeries := epinfo.AllEpisodesInfo(*urlPtr)
	for k, v := range allSeries {
		fmt.Println(strings.Repeat("*", 80))
		fmt.Println(k, ": ")
		for _, epi := range v {
			fmt.Println("Series: ", epi.Series)
			fmt.Println("URL: ", epi.URL)
			fmt.Println("Label: ", epi.Label)
		}
		fmt.Println(strings.Repeat("*", 80))
	}
}
