package gui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/container"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"github.com/gandalf15/iplayerlinks/epinfo"
)

// IPlayerLinksGUI holds all widgets and the window of the GUI
type IPlayerLinksGUI struct {
	sourceURLEnry, allEpURLEntry *widget.Entry
	tvShow, noSeries, noEpisodes *widget.Label
	window                       fyne.Window
	functions                    map[string]func()
	buttons                      map[string]*widget.Button
	checks                       map[string]*widget.Check
	destDir                      string
	allEpURL                     []string
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
	iplGUI.sourceURLEnry = widget.NewEntry()
	iplGUI.sourceURLEnry.SetPlaceHolder("Source iPlayer URL")
	iplGUI.allEpURLEntry = widget.NewMultiLineEntry()
	iplGUI.allEpURLEntry.SetPlaceHolder("Links of all found episodes")
	iplGUI.buttons = make(map[string]*widget.Button)
	iplGUI.functions = make(map[string]func())
	iplGUI.checks = make(map[string]*widget.Check)
	return iplGUI
}

func (iplGUI *IPlayerLinksGUI) saveLinks() {
	f := func(file fyne.URIWriteCloser, e error) {
		defer file.Close()
		file.Write([]byte(iplGUI.allEpURLEntry.Text))
		iplGUI.window.Content().Refresh()
	}
	if len(iplGUI.allEpURLEntry.Text) > 0 {
		dialog.ShowFileSave(f, iplGUI.window)
	} else {
		d := dialog.NewError(errors.New("Nothing to save"), iplGUI.window)
		d.Show()
	}
}

func (iplGUI *IPlayerLinksGUI) getLinks() {
	parsedURL, err := url.Parse(iplGUI.sourceURLEnry.Text)
	if err != nil {
		log.Fatal(err)
	}
	if !strings.Contains(parsedURL.Host, "bbc.co.uk") || !strings.Contains(parsedURL.Path, "iplayer") {
		log.Println("Invalid source URL.")
		d := dialog.NewError(errors.New("Provided source URL is invalid"), iplGUI.window)
		d.Show()
	} else {
		allSeries := epinfo.AllEpisodesInfo(iplGUI.sourceURLEnry.Text, iplGUI.checks["audioDescribed"].Checked,
			iplGUI.checks["signLang"].Checked)
		if len(allSeries) == 0 {
			dialog.NewError(errors.New("No additional links found"), iplGUI.window)
		} else {
			for _, v := range allSeries {
				iplGUI.tvShow.SetText(*v[0].TvShow)
				for _, epi := range v {
					iplGUI.allEpURL = append(iplGUI.allEpURL, epi.URL)
				}
			}
			iplGUI.allEpURLEntry.SetText(strings.Join(iplGUI.allEpURL, "\n"))
			iplGUI.noSeries.SetText(strconv.FormatInt(int64(len(allSeries)), 10))
			iplGUI.noEpisodes.SetText(strconv.FormatInt(int64(len(iplGUI.allEpURL)), 10))

		}

	}
}

func (iplGUI *IPlayerLinksGUI) downloadAllEpisodes() {
	f := func(uri fyne.ListableURI, err error) {
		if err != nil {
			log.Fatalf("Error while opening destination folder %s", err.Error())
		}
		destDir := uri.String()
		iplGUI.destDir = strings.Replace(destDir, "file://", "", 1)
		sub := ""
		if iplGUI.checks["subtitles"].Checked {
			sub = "--all-subs"
		}
		ydl := exec.Command("youtube-dl", "-f", "best", sub, "-o",
			iplGUI.destDir+"/%(title)s-%(release_date)s.%(ext)s", "-a", "-")
		stdin, err := ydl.StdinPipe()
		if err != nil {
			log.Fatalf("Error obtaining stdin: %s", err)
		}
		stdout, err := ydl.StdoutPipe()
		if err != nil {
			log.Fatalf("Error obtaining stdout: %s", err)
		}

		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanRunes)
		go func() {
			f := func() {
				err := errors.New("Try to interupt if not Windows OS")
				if ydl.ProcessState != nil && !ydl.ProcessState.Exited() {
					if runtime.GOOS != "windows" {
						err = ydl.Process.Signal(os.Interrupt)
						time.Sleep(time.Second * 3)
					}
					if err != nil {
						log.Printf("Failed to interrupt youtube-dl process. Error: %s\nKilling the process now...", err)
						err2 := ydl.Process.Kill()
						if err2 != nil {
							d := dialog.NewError(errors.New("Failed to kill youtube-dl process"), iplGUI.window)
							d.Show()
							log.Fatalf("Failed to kill youtube-dl process: %s", err.Error())
						}
					}
				}
			}
			entry := widget.NewMultiLineEntry()
			entry.SetReadOnly(true)
			scrollCont := container.NewScroll(entry)
			scrollCont.SetMinSize(fyne.NewSize(600, 400))
			downloadingCont := container.NewVBox(widget.NewProgressBarInfinite(), scrollCont)
			d := dialog.NewCustom("Downloading", "Cancel", downloadingCont, iplGUI.window)
			d.SetOnClosed(f)
			d.Show()
			textOut := []string{}
			for scanner.Scan() {
				newText := scanner.Text()
				lastNewLine := 0
				if newText == "\r" {
					textOut = textOut[:lastNewLine]
				} else if newText == "\n" {
					lastNewLine = len(textOut)
				}
				textOut = append(textOut, newText)
				entry.SetText(strings.Join(textOut, ""))
				scrollCont.ScrollToBottom()
			}
			d.Hide()
		}()

		go func() {
			defer stdin.Close()
			io.WriteString(stdin, iplGUI.allEpURLEntry.Text)
		}()

		if err := ydl.Start(); nil != err {
			log.Printf("Install youtube-dl first. Failed to start: %s, %s", ydl.Path, err.Error())
			errorDialog := dialog.NewError(fmt.Errorf("Install youtube-dl first. Failed to start: %s, %s",
				ydl.Path, err.Error()), iplGUI.window)
			errorDialog.Show()
		}
		go func() {
			err := ydl.Wait()
			if err != nil {
				log.Printf("Error running the command: '%s' \nError: '%s'", ydl.String(), err)
				errorDialog := dialog.NewError(fmt.Errorf("Error running the command: '%s' \nError: '%s'",
					ydl.String(), err), iplGUI.window)
				errorDialog.Show()
			} else {
				successDialog := dialog.NewInformation("Finished", "Success", iplGUI.window)
				successDialog.Show()
			}
		}()
	}
	if len(iplGUI.allEpURLEntry.Text) > 0 {
		d := dialog.NewFolderOpen(f, iplGUI.window)
		d.Show()
	} else {
		d := dialog.NewError(errors.New("Nothing to download"), iplGUI.window)
		d.Show()
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
	iplGUI.functions["saveLinks"] = func() { iplGUI.saveLinks() }
	iplGUI.buttons["saveLinks"].OnTapped = iplGUI.functions["saveLinks"]

	iplGUI.buttons["getLinks"] = widget.NewButton("Get Links", func() {})
	iplGUI.functions["getLinks"] = func() { iplGUI.getLinks() }
	iplGUI.buttons["getLinks"].OnTapped = iplGUI.functions["getLinks"]

	iplGUI.functions["downloadAll"] = func() { iplGUI.downloadAllEpisodes() }
	iplGUI.buttons["downloadAll"] = widget.NewButton("Download All Episodes", iplGUI.functions["downloadAll"])

	iplGUI.checks["audioDescribed"] = widget.NewCheck("Audio Described Links", func(bool) {})
	iplGUI.checks["signLang"] = widget.NewCheck("Sign Language Links", func(bool) {})
	iplGUI.checks["subtitles"] = widget.NewCheck("Download Subtitles", func(bool) {})

	subtitleCont := container.NewHBox(layout.NewSpacer(), iplGUI.checks["subtitles"], layout.NewSpacer())
	bottomContainer := container.NewVBox(iplGUI.buttons["saveLinks"], subtitleCont, iplGUI.buttons["downloadAll"], statusBar)
	allSeriesContainer := container.NewScroll(iplGUI.allEpURLEntry)
	checksContainer := container.NewHBox(iplGUI.checks["audioDescribed"], layout.NewSpacer(), iplGUI.checks["signLang"])
	topContainer := container.NewVBox(container.NewScroll(iplGUI.sourceURLEnry), checksContainer,
		iplGUI.buttons["getLinks"])
	content := container.NewBorder(topContainer, bottomContainer, nil, nil, allSeriesContainer)
	iplGUI.window.Resize(fyne.NewSize(800, 600))
	iplGUI.window.CenterOnScreen()
	iplGUI.window.SetContent(content)
	iplGUI.window.Canvas().SetOnTypedKey(iplGUI.typedKey)
	iplGUI.window.ShowAndRun()
}
