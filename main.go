package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/inputs"
	catppuccin "github.com/mbaklor/fyne-catppuccin"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var inputName = "LikeAlertText"
var mu sync.Mutex
var stopPolling = false

func getLikeCount(apiKey, videoID string) (uint64, error) {
	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return 0, err
	}

	call := service.Videos.List([]string{"statistics"}).Id(videoID)
	response, err := call.Do()
	if err != nil {
		return 0, err
	}

	if len(response.Items) > 0 {
		stats := response.Items[0].Statistics
		return stats.LikeCount, nil
	}
	return 0, fmt.Errorf("no video found")
}

func startPolling(apiKey, videoID string, label *widget.Label) {
	go func() {
		var lastCount uint64 = 0

		client, err := goobs.New("localhost:4455", goobs.WithPassword("YOUR_PASSWORD"))
		if err != nil {
			log.Fatal(err)
		}

		for {
			mu.Lock()
			if stopPolling {
				mu.Unlock()
				break
			}
			mu.Unlock()

			count, err := getLikeCount(apiKey, videoID)
			if err != nil {
				label.SetText("Error: " + err.Error())
				time.Sleep(30 * time.Second)
				continue
			}

			if count != lastCount {
				label.SetText("Likes: " + strconv.FormatUint(count, 10))
				lastCount = count
				_, err = client.Inputs.SetInputSettings(&inputs.SetInputSettingsParams{
					InputName: &inputName,
					InputSettings: map[string]interface{}{
						"text": fmt.Sprintf("👍 Like count: %d", count),
					},
				})
				if err != nil {
					log.Print(err)
				}
			}
			time.Sleep(15 * time.Second)
		}
	}()
}

func main() {
	a := app.New()
	w := a.NewWindow("YouTube Like Monitor")

	ctp := catppuccin.New()
	ctp.SetFlavor(catppuccin.Latte)
	a.Settings().SetTheme(ctp)

	// UI Elements
	ApiKeyEntry := widget.NewEntry()
	ApiKeyEntry.SetPlaceHolder("Enter YouTube API key")

	videoIDEntry := widget.NewEntry()
	videoIDEntry.SetPlaceHolder("Enter YouTube Video ID")

	likeLabel := widget.NewLabel("Likes: N/A")

	startButton := widget.NewButton("Start", func() {
		apiKey := strings.TrimSpace(ApiKeyEntry.Text)
		if apiKey == "" {
			likeLabel.SetText("Please enter a Video ID")
			return
		}

		videoID := strings.TrimSpace(videoIDEntry.Text)
		if videoID == "" {
			likeLabel.SetText("Please enter a Video ID")
			return
		}
		mu.Lock()
		stopPolling = false
		mu.Unlock()
		go startPolling(apiKey, videoID, likeLabel)
	})

	stopButton := widget.NewButton("Stop", func() {
		mu.Lock()
		stopPolling = true
		mu.Unlock()
	})

	mainTab := container.NewVBox(
		widget.NewLabel("YouTube Video API key:"),
		ApiKeyEntry,
		widget.NewLabel("YouTube Video ID:"),
		videoIDEntry,
		container.NewHBox(startButton, stopButton),
		likeLabel,
	)

	helpText := `🎥 HOW TO FIND A VIDEO ID:
- From a YouTube link like: https://www.youtube.com/watch?v=ABC123XYZ
- Copy only the part after v= (e.g. ABC123XYZ)

🔑 HOW TO GET A YOUTUBE API KEY:
1. Visit https://console.cloud.google.com/
2. Create a project
3. Go to APIs & Services → Library
4. Search for "YouTube Data API v3" → Enable it
5. Go to APIs & Services → Credentials
6. Create an API Key and copy it

⚠️ Quota Note:
- YouTube API has daily limits. Don’t poll too frequently.`

	helpTab := container.NewScroll(widget.NewLabel(helpText))

	tabs := container.NewAppTabs(
		container.NewTabItem("Monitor", mainTab),
		container.NewTabItem("Help", helpTab),
	)

	w.SetContent(tabs)
	w.Resize(fyne.NewSize(500, 300))
	w.ShowAndRun()
}
