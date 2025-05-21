package rod

import (
	"context"
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
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
	tabLimit        = 5
	openPages       []*rod.Page
	idleTimeout     = 2 * time.Minute
	idleTimer       *time.Timer
	timerMu         sync.Mutex
	initErr         error
	initOnce        sync.Once
	headless        = true
)

func SetHeadless(h bool) {
	headless = h
}

func SetTabLimit(limit int) {
	tabLimit = limit
}

func SetIdleTimeout(timeout time.Duration) {
	idleTimeout = timeout
}

func InitBrowser() (*rod.Browser, error) {
	initOnce.Do(func() {
		browserMu.Lock()
		defer browserMu.Unlock()
		if err := startBrowser(); err != nil {
			initErr = err
		}
		startIdleTimer()
	})

	browserMu.Lock()
	defer browserMu.Unlock()
	return browserInstance, initErr
}

func startBrowser() error {
	if browserInstance != nil {
		cleanupBrowserResources()
	}

	l := launcher.New().Headless(headless)
	url, err := l.Launch()
	if err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}

	browser := rod.New().ControlURL(url).NoDefaultDevice()
	if err := browser.Connect(); err != nil {
		l.Kill()
		return fmt.Errorf("failed to connect to browser: %w", err)
	}

	browserInstance = browser
	openPages = nil
	return nil
}

func startIdleTimer() {
	timerMu.Lock()
	defer timerMu.Unlock()

	if idleTimer != nil {
		idleTimer.Stop()
	}

	idleTimer = time.AfterFunc(idleTimeout, func() {
		RestartBrowser()
	})
}

func resetIdleTimer() {
	timerMu.Lock()
	defer timerMu.Unlock()

	if idleTimer != nil {
		idleTimer.Reset(idleTimeout)
	}
}

func RestartBrowser() {
	browserMu.Lock()
	defer browserMu.Unlock()

	if browserInstance != nil {
		cleanupBrowserResources()
	}

	if err := startBrowser(); err != nil {
		initErr = err
		return
	}
	startIdleTimer()
}

func cleanupBrowserResources() {
	for _, p := range openPages {
		p.Close()
	}
	openPages = nil

	if err := browserInstance.Close(); err != nil {
		logrus.Warnf("Error closing browser: %v", err)
	}
	browserInstance = nil
}

func getNewPage() (*rod.Page, error) {
	browserMu.Lock()
	defer browserMu.Unlock()

	if browserInstance == nil {
		return nil, fmt.Errorf("browser not initialized")
	}

	if len(openPages) >= tabLimit && tabLimit > 0 {
		oldest := openPages[0]
		oldest.Close()
		openPages = openPages[1:]
	}

	page, err := browserInstance.Page(proto.TargetCreateTarget{})
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}
	openPages = append(openPages, page)
	return page, nil
}

func ClosePage(page *rod.Page) {
	if page == nil {
		return
	}

	page.Close()

	browserMu.Lock()
	defer browserMu.Unlock()

	for i, p := range openPages {
		if p == page {
			openPages = append(openPages[:i], openPages[i+1:]...)
			break
		}
	}
}

func CaptureScreenshot(ctx context.Context, browser *rod.Browser, opt ScreenshotOptions) ([]byte, error) {
	_, err := InitBrowser()
	if err != nil {
		return nil, fmt.Errorf("browser initialization failed: %w", err)
	}

	resetIdleTimer()

	page, err := getNewPage()
	if err != nil {
		return nil, fmt.Errorf("failed to get page: %w", err)
	}
	defer ClosePage(page)

	page = page.Context(ctx)

	if err := page.Navigate(opt.URL); err != nil {
		return nil, fmt.Errorf("navigation failed: %w", err)
	}

	if err := page.WaitLoad(); err != nil {
		return nil, fmt.Errorf("page load failed: %w", err)
	}

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

			logrus.Info("Emulating custom mobile device")
			if err := page.Emulate(device); err != nil {
				logrus.Error("failed to emulate custom mobile: ", err)
				return nil, fmt.Errorf("failed to emulate custom mobile: %w", err)
			}

			logrus.Info("Emulate custom mobile complete")
		} else {
			logrus.Info("Emulating default mobile device (iPhoneX)")
			if err := page.Emulate(devices.IPhoneX); err != nil {
				logrus.Error("failed to emulate default mobile: ", err)
				return nil, fmt.Errorf("failed to emulate default mobile: %w", err)
			}
			logrus.Info("Emulate default mobile complete")
		}
	} else if opt.Width > 0 && opt.Height > 0 {
		logrus.Infof("Setting viewport to %dx%d", opt.Width, opt.Height)
		page.MustSetViewport(opt.Width, opt.Height, 1.0, false)
		logrus.Info("Viewport set")
	}

	logrus.Info("Waiting for page load after emulation/viewport")
	page.MustWaitLoad()
	logrus.Info("Page loaded after emulation/viewport")

	if opt.DelaySeconds > 0 {
		logrus.Infof("Waiting for %d seconds", opt.DelaySeconds)
		time.Sleep(time.Duration(opt.DelaySeconds) * time.Second)
		logrus.Info("Delay complete")
	}

	var buf []byte
	var errors error

	if opt.FullPage {
		logrus.Info("Taking full page screenshot")
		buf, errors = page.Screenshot(true, nil)
	} else {
		logrus.Info("Taking viewport screenshot")
		buf, errors = page.Screenshot(false, nil)
	}

	if errors != nil {
		logrus.Error("failed to take screenshot: ", errors)
		return nil, fmt.Errorf("failed to take screenshot: %w", errors)
	}

	logrus.Info("Screenshot taken successfully")
	return buf, nil
}
