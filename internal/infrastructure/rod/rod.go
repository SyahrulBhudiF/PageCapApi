package rod

import (
	"context"
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type ScreenshotOptions struct {
	URL          string
	Width        int
	Height       int
	FullPage     bool
	DelaySeconds int
	IsMobile     bool
}

var (
	browserInstance *rod.Browser
	browserMu       sync.Mutex

	tabLimit    = 5
	openPages   []*rod.Page
	idleTimeout = 2 * time.Minute

	idleTimer *time.Timer
	timerMu   sync.Mutex
	initErr   error
	once      sync.Once
	headless  = true
)

// InitBrowser initializes or returns the singleton browser instance.
func InitBrowser() (*rod.Browser, error) {
	once.Do(func() {
		err := startBrowser()
		if err != nil {
			initErr = err
			return
		}
		startIdleTimer()
	})
	return browserInstance, initErr
}

func startBrowser() error {
	l := launcher.New()
	if headless {
		l = l.Headless(true)
	}
	url, err := l.Launch()
	if err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}

	browser := rod.New().ControlURL(url).NoDefaultDevice()
	if err := browser.Connect(); err != nil {
		return fmt.Errorf("failed to connect to browser: %w", err)
	}

	logrus.Info("Rod browser initialized")

	browserMu.Lock()
	defer browserMu.Unlock()
	browserInstance = browser
	openPages = nil // reset tabs
	return nil
}

// startIdleTimer starts the idle timeout timer that will restart the browser after inactivity.
func startIdleTimer() {
	timerMu.Lock()
	defer timerMu.Unlock()
	if idleTimer != nil {
		idleTimer.Stop()
	}
	idleTimer = time.AfterFunc(idleTimeout, func() {
		logrus.Info("Idle timeout reached. Restarting browser...")
		restartBrowser()
	})
}

// resetIdleTimer resets the idle timer whenever there is activity.
func resetIdleTimer() {
	timerMu.Lock()
	defer timerMu.Unlock()
	if idleTimer != nil {
		idleTimer.Reset(idleTimeout)
	}
}

// restartBrowser closes the old browser and starts a new one.
func restartBrowser() {
	browserMu.Lock()
	defer browserMu.Unlock()

	if browserInstance != nil {
		// Close all open pages
		for _, p := range openPages {
			_ = p.Close()
		}
		openPages = nil

		// Close browser
		_ = browserInstance.Close()
		browserInstance = nil
	}

	if err := startBrowser(); err != nil {
		logrus.Errorf("Failed to restart browser: %v", err)
		return
	}
	startIdleTimer()
}

// getNewPage returns a new page, closing the oldest page if tabLimit exceeded.
func getNewPage() (*rod.Page, error) {
	browserMu.Lock()
	defer browserMu.Unlock()

	if browserInstance == nil {
		if err := startBrowser(); err != nil {
			return nil, err
		}
		startIdleTimer()
	}

	// Close oldest tab if limit exceeded
	if len(openPages) >= tabLimit {
		oldest := openPages[0]
		if err := oldest.Close(); err != nil {
			logrus.Warnf("Failed to close oldest page: %v", err)
		}
		openPages = openPages[1:]
	}

	page := browserInstance.MustPage()
	openPages = append(openPages, page)
	return page, nil
}

func CaptureScreenshot(ctx context.Context, browser *rod.Browser, opt ScreenshotOptions) ([]byte, error) {
	resetIdleTimer()

	page, errors := getNewPage()
	if errors != nil {
		return nil, fmt.Errorf("failed to get new page: %w", errors)
	}

	page.MustNavigate(opt.URL)
	page.MustWaitLoad()

	if opt.IsMobile {
		if opt.Width > 0 && opt.Height > 0 {
			device := devices.Device{
				Title: "Custom Mobile",
				Screen: struct {
					DevicePixelRatio float64
					Horizontal       devices.ScreenSize
					Vertical         devices.ScreenSize
				}{
					DevicePixelRatio: 2.0,
					Horizontal:       devices.ScreenSize{},
					Vertical: devices.ScreenSize{
						Width:  opt.Width,
						Height: opt.Height,
					},
				},
				UserAgent: devices.IPhoneX.UserAgent,
			}

			if err := page.Emulate(device); err != nil {
				logrus.Error("failed to emulate custom mobile: ", err)
				return nil, fmt.Errorf("failed to emulate custom mobile: %w", err)
			}

			logrus.Info("Emulate custom mobile")
		} else {
			if err := page.Emulate(devices.IPhoneX); err != nil {
				logrus.Error("failed to emulate default mobile: ", err)
				return nil, fmt.Errorf("failed to emulate default mobile: %w", err)
			}
			logrus.Info("Emulate default mobile")
		}
	} else if opt.Width > 0 && opt.Height > 0 {
		logrus.Info("Set viewport")
		page.MustSetViewport(opt.Width, opt.Height, 1.0, false)
	}

	page.MustWaitLoad()

	if opt.DelaySeconds > 0 {
		logrus.Infof("Waiting for %d seconds", opt.DelaySeconds)
		time.Sleep(time.Duration(opt.DelaySeconds) * time.Second)
	}

	var buf []byte
	var err error

	if opt.FullPage {
		logrus.Info("Taking full page screenshot")
		buf, err = page.Screenshot(true, nil)
	} else {
		logrus.Info("Taking viewport screenshot")
		buf, err = page.Screenshot(false, nil)
	}

	if err != nil {
		logrus.Error("failed to take screenshot: ", err)
		return nil, fmt.Errorf("failed to take screenshot: %w", err)
	}

	logrus.Info("Screenshot taken successfully")
	return buf, nil
}
