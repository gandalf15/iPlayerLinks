package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/container"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"github.com/gandalf15/iplayer/epinfo"
)

type iPlayerLinks struct {
	sourceURL            string
	allSeries            map[string][]epinfo.EpisodeInfo
	noSeries, noEpisodes widget.Label
	window               fyne.Window
}

func main() {
	urlPtr := flag.String("url", "", "-url=[iPlayer URL with episodes]")
	flag.Parse()
	if *urlPtr == "" {
		myApp := app.New()
		ipl := iPlayerLinks{}
		ipl.window = myApp.NewWindow("iPlayerLinks")
		sourceURL := widget.NewEntry()
		sourceURL.SetPlaceHolder("Source iPlayer URL")
		allSeries := widget.NewMultiLineEntry()
		statusBar := widget.NewHBox(widget.NewLabel("No. Of Series:"), &ipl.noSeries,
			layout.NewSpacer(), widget.NewLabel("No. Of All Ep:"), &ipl.noEpisodes)
		button := widget.NewButton("Get All Links", func() {
			ipl.sourceURL = sourceURL.Text
			ipl.allSeries = epinfo.AllEpisodesInfo(ipl.sourceURL)
			outString := ""
			epCounter := 0
			for _, v := range ipl.allSeries {
				// outString += k + "\n"
				for _, epi := range v {
					// outString += "\nSeries: " + epi.Series
					// outString += "\nURL: " + epi.URL
					outString += epi.URL + "\n"
					epCounter++
					// outString += "\nLabel: " + epi.Label
				}
			}
			allSeries.SetText(outString)
			ipl.noSeries.SetText(strconv.FormatInt(int64(len(ipl.allSeries)), 10))
			ipl.noEpisodes.SetText(strconv.FormatInt(int64(epCounter), 10))
		})
		allSeriesContainer := container.NewScroll(allSeries)
		topContainer := container.NewVBox(container.NewScroll(sourceURL), button)
		content := container.NewBorder(topContainer, statusBar, nil, nil, allSeriesContainer)
		ipl.window.Resize(fyne.NewSize(500, 500))
		ipl.window.CenterOnScreen()
		ipl.window.SetContent(content)
		ipl.window.ShowAndRun()
		// log.Fatal("usage: ./iplayer -url=[iPlayer URL with episodes]")
	} else {
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

}
