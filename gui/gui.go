package gui

import (
	"errors"
	"log"
	"net/url"
	"strconv"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/container"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"github.com/gandalf15/iplayer/epinfo"
)

// IPlayerLinksGUI holds all widgets and the window of the GUI
type IPlayerLinksGUI struct {
	sourceURL, allEpisodesURL    *widget.Entry
	tvShow, noSeries, noEpisodes *widget.Label
	window                       fyne.Window
	functions                    map[string]func()
	buttons                      map[string]*widget.Button
	checks                       map[string]*widget.Check
}

func (iplGUI *IPlayerLinksGUI) addButton(text string, action func()) *widget.Button {
	button := widget.NewButton(text, action)
	iplGUI.buttons[text] = button
	return button
}

func (iplGUI *IPlayerLinksGUI) typedKey(ev *fyne.KeyEvent) {
	if ev.Name == fyne.KeyReturn || ev.Name == fyne.KeyEnter {
		action := iplGUI.functions["getAllLinks"]
		action()
	}
}

// NewIplayerLinksGUI initialises and returns a pointer to iPlayerLinksGUI
func NewIplayerLinksGUI(myApp fyne.App) *IPlayerLinksGUI {
	iplGUI := &IPlayerLinksGUI{}
	iplGUI.noSeries = widget.NewLabel("0")
	iplGUI.noEpisodes = widget.NewLabel("0")
	iplGUI.tvShow = widget.NewLabel("")
	iplGUI.window = myApp.NewWindow("iPlayerLinks")
	iplGUI.sourceURL = widget.NewEntry()
	iplGUI.sourceURL.SetPlaceHolder("Source iPlayer URL")
	iplGUI.allEpisodesURL = widget.NewMultiLineEntry()
	iplGUI.allEpisodesURL.SetReadOnly(true)
	iplGUI.allEpisodesURL.SetPlaceHolder("Links of all found episodes")
	iplGUI.buttons = make(map[string]*widget.Button)
	iplGUI.functions = make(map[string]func())
	iplGUI.checks = make(map[string]*widget.Check)
	return iplGUI
}

func (iplGUI *IPlayerLinksGUI) saveLinks() {
	f := func(file fyne.URIWriteCloser, e error) {
		defer file.Close()
		file.Write([]byte(iplGUI.allEpisodesURL.Text))
		iplGUI.buttons["saveLinks"].Importance = widget.MediumImportance
	}
	dialog.ShowFileSave(f, iplGUI.window)
}

func (iplGUI *IPlayerLinksGUI) getLinks() {
	parsedURL, err := url.Parse(iplGUI.sourceURL.Text)
	if err != nil {
		log.Fatal(err)
	}
	if !strings.Contains(parsedURL.Host, "bbc.co.uk") || !strings.Contains(parsedURL.Path, "iplayer") {
		log.Println("Invalid source URL.")
		d := dialog.NewError(errors.New("Provided source URL is invalid"), iplGUI.window)
		d.Show()
	} else {
		allSeries := epinfo.AllEpisodesInfo(iplGUI.sourceURL.Text, iplGUI.checks["audioDescribed"].Checked,
			iplGUI.checks["signLang"].Checked)
		if len(allSeries) == 0 {
			dialog.NewError(errors.New("No additional links found"), iplGUI.window)
		} else {
			var epLinks []string
			for _, v := range allSeries {
				iplGUI.tvShow.SetText(*v[0].TvShow)
				for _, epi := range v {
					epLinks = append(epLinks, epi.URL)
				}
			}
			iplGUI.allEpisodesURL.SetText(strings.Join(epLinks, "\n"))
			iplGUI.noSeries.SetText(strconv.FormatInt(int64(len(allSeries)), 10))
			iplGUI.noEpisodes.SetText(strconv.FormatInt(int64(len(epLinks)), 10))
			iplGUI.allEpisodesURL.SetReadOnly(false)
			iplGUI.buttons["saveLinks"].Enable()
			iplGUI.buttons["saveLinks"].Importance = widget.HighImportance
			iplGUI.buttons["getLinks"].Importance = widget.MediumImportance
		}

	}
}

// Gui creates and shows GUI for iPlayerLinks
func Gui() {
	myApp := app.New()
	myApp.SetIcon(resourceIconPng)
	iplGUI := NewIplayerLinksGUI(myApp)
	statusBar := widget.NewHBox(widget.NewLabel("No. Of Series:"), iplGUI.noSeries,
		layout.NewSpacer(), iplGUI.tvShow, layout.NewSpacer(),
		widget.NewLabel("No. Of All Ep:"), iplGUI.noEpisodes)

	iplGUI.buttons["saveLinks"] = widget.NewButton("Save Links To File", func() {})
	iplGUI.buttons["saveLinks"].Disable()
	iplGUI.buttons["getLinks"] = widget.NewButton("Get Links", func() {})
	iplGUI.buttons["getLinks"].Importance = widget.HighImportance
	iplGUI.functions["saveLinks"] = func() { iplGUI.saveLinks() }
	iplGUI.buttons["saveLinks"].OnTapped = iplGUI.functions["saveLinks"]
	iplGUI.functions["getLinks"] = func() { iplGUI.getLinks() }
	iplGUI.buttons["getLinks"].OnTapped = iplGUI.functions["getLinks"]
	iplGUI.checks["audioDescribed"] = widget.NewCheck("Audio Described Links", func(bool) {})
	iplGUI.checks["signLang"] = widget.NewCheck("Sign Language Links", func(bool) {})

	bottomContainer := container.NewVBox(iplGUI.buttons["saveLinks"], statusBar)
	allSeriesContainer := container.NewScroll(iplGUI.allEpisodesURL)
	checksContainer := container.NewHBox(iplGUI.checks["audioDescribed"], layout.NewSpacer(), iplGUI.checks["signLang"])
	topContainer := container.NewVBox(container.NewScroll(iplGUI.sourceURL), checksContainer,
		iplGUI.buttons["getLinks"])
	content := container.NewBorder(topContainer, bottomContainer, nil, nil, allSeriesContainer)
	iplGUI.window.Resize(fyne.NewSize(800, 600))
	iplGUI.window.CenterOnScreen()
	iplGUI.window.SetContent(content)
	iplGUI.window.Canvas().SetOnTypedKey(iplGUI.typedKey)
	iplGUI.window.ShowAndRun()
}
