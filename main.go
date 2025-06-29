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
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/inputs"
	catppuccin "github.com/mbaklor/fyne-catppuccin"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var (
	inputName    = "LikeAlertText"
	pollingMutex sync.Mutex
	stopPolling  = false
)

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

	if len(response.Items) == 0 {
		return 0, fmt.Errorf("no video found")
	}

	return response.Items[0].Statistics.LikeCount, nil
}

func startPolling(apiKey, videoID, obsWsURL, obsWsPassword string, label *widget.Label, errorText *canvas.Text, startBtn, stopBtn *widget.Button) {
	go func() {
		var lastCount uint64

		client, err := goobs.New(obsWsURL, goobs.WithPassword(obsWsPassword))
		if err != nil {
			errorText.Text = "❌ OBS connection failed"
			log.Println("OBS error:", err)

			startBtn.Enable()
			stopBtn.Disable()
			return
		}

		for {
			pollingMutex.Lock()
			if stopPolling {
				pollingMutex.Unlock()
				break
			}
			pollingMutex.Unlock()

			count, err := getLikeCount(apiKey, videoID)
			if err != nil {
				errorText.Text = "❌ YouTube error: " + err.Error()
				time.Sleep(30 * time.Second)
				continue
			}

			if count != lastCount {
				lastCount = count
				label.SetText("👍 Likes: " + strconv.FormatUint(count, 10))

				_, err = client.Inputs.SetInputSettings(&inputs.SetInputSettingsParams{
					InputName: &inputName,
					InputSettings: map[string]interface{}{
						"likes": fmt.Sprintf("%d", count),
					},
				})
				if err != nil {
					errorText.Text = fmt.Sprintf("OBS update error:", err)
				}
			}

			time.Sleep(15 * time.Second)
		}

		// Re-enable UI buttons after stop
		startBtn.Enable()
		stopBtn.Disable()
	}()
}

func main() {
	a := app.New()
	w := a.NewWindow("YouTube Like Monitor")

	// Apply Catppuccin theme
	ctp := catppuccin.New()
	ctp.SetFlavor(catppuccin.Macchiato)
	a.Settings().SetTheme(ctp)

	// UI elements
	apiKeyEntry := widget.NewEntry()
	apiKeyEntry.SetPlaceHolder("YouTube API Key")

	videoIDEntry := widget.NewEntry()
	videoIDEntry.SetPlaceHolder("e.g. dQw4w9WgXcQ")

	obsWsEntry := widget.NewEntry()
	obsWsEntry.SetPlaceHolder("localhost:4455")

	obsPasswordEntry := widget.NewPasswordEntry()
	obsPasswordEntry.SetPlaceHolder("optional")

	likeLabel := widget.NewLabel("Likes: N/A")

	errorText := canvas.NewText("", theme.Color(theme.ColorNameError))
	errorText.TextStyle = fyne.TextStyle{Bold: true}

	startButton := widget.NewButtonWithIcon("Start", theme.MediaPlayIcon(), nil)
	stopButton := widget.NewButtonWithIcon("Stop", theme.MediaStopIcon(), nil)

	// Initially stop is disabled
	stopButton.Disable()

	startButton.OnTapped = func() {
		apiKey := strings.TrimSpace(apiKeyEntry.Text)
		videoID := strings.TrimSpace(videoIDEntry.Text)

		if apiKey == "" || videoID == "" {
			errorText.Text = "You must provide both a YouTube API key and Video ID."
			return
		}

		obsWsURL := strings.TrimSpace(obsWsEntry.Text)
		if obsWsURL == "" {
			obsWsURL = "localhost:4455"
		}
		obsPassword := strings.TrimSpace(obsPasswordEntry.Text)

		// Set flags and button states
		pollingMutex.Lock()
		stopPolling = false
		startButton.Disable()
		stopButton.Enable()
		pollingMutex.Unlock()

		startPolling(apiKey, videoID, obsWsURL, obsPassword, likeLabel, errorText, startButton, stopButton)
		errorText.Text = ""
	}

	stopButton.OnTapped = func() {
		pollingMutex.Lock()
		stopPolling = true
		startButton.Enable()
		stopButton.Disable()
		pollingMutex.Unlock()
	}

	// Layout
	form := container.NewVBox(
		widget.NewLabel("🔑 YouTube API Key"),
		apiKeyEntry,
		widget.NewLabel("🎥 YouTube Video ID"),
		videoIDEntry,
		widget.NewLabel("📡 OBS WebSocket URL"),
		obsWsEntry,
		widget.NewLabel("🔐 OBS WebSocket Password"),
		obsPasswordEntry,
		startButton,
		stopButton,
		layout.NewSpacer(),
		likeLabel,
		container.NewPadded(errorText),
	)

	helpText := `🎥 HOW TO FIND A VIDEO ID:
- Copy the part after v= in the YouTube URL: https://youtube.com/watch?v=ABC123 → ABC123

🔑 HOW TO GET A YOUTUBE API KEY:
1. Visit https://console.cloud.google.com/
2. Create a project
3. Go to APIs & Services → Library
4. Search for "YouTube Data API v3" and enable it
5. Go to Credentials → Create API Key

⚠️ Usage Tip:
- Don’t poll too often. The API has quota limits.`

	helpTab := container.NewScroll(widget.NewLabel(helpText))

	tabs := container.NewAppTabs(
		container.NewTabItem("Monitor", form),
		container.NewTabItem("Help", helpTab),
	)

	w.SetContent(tabs)
	w.Resize(fyne.NewSize(500, 500))
	w.ShowAndRun()
}
