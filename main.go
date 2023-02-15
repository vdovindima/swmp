package main

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/dhowden/tag"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type track struct {
	partNumber int
	title      string
	artist     string
	fileName   string
}

var (
	data   = []track{}
	folder = "./audio/"
)

func playSong(s beep.StreamSeekCloser, format beep.Format) {
	done := make(chan bool)
	// Play the audio stream
	speaker.Play(beep.Seq(s, beep.Callback(func() {
		done <- true
	})))
	<-done
}

func executorSong(fileName chan string) {
	var once sync.Once
	for {
		name := <-fileName
		log.Println(name)
		// Open the audio file
		file, err := os.Open(name)
		if err != nil {
			panic(err)
		}

		// Decode the audio file
		streamer, format, err := mp3.Decode(file)
		if err != nil {
			panic(err)
		}

		onceBody := func() {
			// Initialize the speaker with the format
			speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		}

		once.Do(onceBody)
		speaker.Clear()
		go playSong(streamer, format)
	}
}

func main() {
	fileName := make(chan string)
	go executorSong(fileName)

	myApp := app.New()
	myWindow := myApp.NewWindow("List Widget")
	myApp.Settings().SetTheme(theme.DefaultTheme())

	files, err := ioutil.ReadDir(folder)
	if err != nil {
		log.Println(err)
	}

	for i := 0; i < len(files); i++ {
		if !files[i].IsDir() {
			date, err := os.Open(folder + files[i].Name())
			if err != nil {
				log.Println(err)
			}
			m, err := tag.ReadFrom(date)
			if err != nil {
				log.Println(err)
			}
			if m.Artist() != "" && m.Title() != "" {
				data = append(data, track{partNumber: i, title: m.Title(), artist: m.Artist(), fileName: files[i].Name()})
			}
		}
	}

	list := widget.NewList(
		func() int {
			return len(data)
		},
		func() fyne.CanvasObject {
			return widget.NewButton("click me", func() {})
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Button).SetText(data[i].artist + " - " + data[i].title)
			o.(*widget.Button).SetIcon(theme.AccountIcon())
			o.(*widget.Button).OnTapped = func() {
				fileName <- folder + data[i].fileName
			}
		},
	)
	myWindow.SetContent(list)
	myWindow.Resize(fyne.NewSize(460, 460))
	myWindow.ShowAndRun()
}
