package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/gandalf15/iplayer/epinfo"
)

func main() {
	urlPtr := flag.String("url", "", "iPlayer URL with all episodes")
	flag.Parse()
	if *urlPtr == "" {
		log.Fatal("You must provide iPlayer URL with all episodes. I got: ", *urlPtr)
	}
	allEP := epinfo.AllEpisodesInfo(*urlPtr)
	for k, v := range allEP {
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
