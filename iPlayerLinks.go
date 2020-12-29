package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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
		tvShow := ""
		outString := ""
		saveLinksButton := widget.NewButton("Save All Links To File", func() {})
		f := func() {
			f, err := os.Create(tvShow + ".txt")

			if err != nil {
				log.Fatal(err)
			}

			defer f.Close()

			_, err2 := f.WriteString(outString)

			if err2 != nil {
				log.Fatal(err2)
			}
			saveLinksButton.Disable()
			saveLinksButton.SetText("Saved")
		}
		saveLinksButton.OnTapped = f
		saveLinksButton.Disable()
		getLinksButton := widget.NewButton("Get All Links", func() {
			ipl.sourceURL = sourceURL.Text
			ipl.allSeries = epinfo.AllEpisodesInfo(ipl.sourceURL)
			epCounter := 0
			for _, v := range ipl.allSeries {
				tvShow = *v[0].TvShow
				for _, epi := range v {
					outString += epi.URL + "\n"
					epCounter++
				}
			}
			allSeries.SetText(outString)
			ipl.noSeries.SetText(strconv.FormatInt(int64(len(ipl.allSeries)), 10))
			ipl.noEpisodes.SetText(strconv.FormatInt(int64(epCounter), 10))
			saveLinksButton.SetText("Save All Links To File")
			saveLinksButton.Enable()
		})
		allSeriesContainer := container.NewScroll(allSeries)
		topContainer := container.NewVBox(container.NewScroll(sourceURL), getLinksButton, saveLinksButton)
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
