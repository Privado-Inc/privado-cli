package utils

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/schollz/progressbar/v3"
)

func OpenURLInBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = nil
	}

	if cmd != nil {
		if err := cmd.Start(); err == nil {
			return
		}
	}

	// in case we cannot automatically open due to
	// unknown OS or an error, print
	fmt.Println("\n> Unable to open browser")
	fmt.Println("> Open the following URL to view results:", url)
}

// Ignores all error and waits for URL to be responsive
// by sending HEAD request every intervalSeconds
func WaitForResponsiveURL(url string, intervalSeconds int) {
	if intervalSeconds == 0 {
		intervalSeconds = 5
	}
	ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		res, err := http.Head(url)
		if err == nil && res.StatusCode == 200 {
			return
		}
	}
}

func WaitAndOpenURL(url string, sgn chan bool, interval int) {
	WaitForResponsiveURL(url, interval)
	sgn <- true
	OpenURLInBrowser(url)
}

func RunOnCtrlC(cleanupFn func()) chan os.Signal {
	notifySignal := make(chan os.Signal)
	signal.Notify(notifySignal, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-notifySignal
		cleanupFn()
		os.Exit(0)
	}()

	return notifySignal
}

func ClearSignals(sgn chan os.Signal) {
	signal.Stop(sgn)
}

func RenderProgressSpinnerWithMessages(complete, quit chan bool, loadMessages, afterLoadMessages []string) {
	if len(loadMessages) == 0 {
		// default message
		loadMessages = []string{"Loading.."}
	}

	messageIndex := 0
	messageRotationTicker := time.NewTicker(20 * time.Second)

	bar := progressbar.NewOptions(-1,
		progressbar.OptionFullWidth(),
		// Good spinners: 0, 31, 51, 52, 54
		progressbar.OptionSpinnerType(52),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetDescription(loadMessages[messageIndex]),
		progressbar.OptionShowCount(),
	)

	for {
		seconds := bar.State().SecondsSince
		select {
		case <-quit:
			fmt.Println()
			bar.Close()
		case <-complete:
			bar.Close()
			fmt.Println()
			fmt.Println("> Complete")
			fmt.Println("> Total Time taken:", seconds, "seconds")

			if len(afterLoadMessages) > 0 {
				for _, message := range afterLoadMessages {
					fmt.Print("\n> ", message)
				}
			}
			fmt.Println()
			return
		case <-messageRotationTicker.C:
			messageIndex++
			bar.Describe(loadMessages[messageIndex%len(loadMessages)])
		default:
			bar.Set(int(seconds))
			time.Sleep(150 * time.Millisecond)
		}
	}
}

func ShowConfirmationPrompt(msg string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s (y/N): ", msg)
	ans, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	ans = strings.TrimSpace(ans)
	ans = strings.ToLower(ans)

	if ans == "y" || ans == "yes" || ans == "1" {
		return true, nil
	}
	return false, nil
}
