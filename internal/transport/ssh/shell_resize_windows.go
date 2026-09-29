//go:build windows

package ssh

import (
	"time"

	"github.com/sony/niwashi/internal/logger"
	"golang.org/x/sys/windows"
)

const refreshInterval = 100 * time.Millisecond

func handleShellResize(s *sshSession) (func(), error) {
	h, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
	if err != nil {
		return nil, err
	}

	var size windows.Coord
	done := make(chan struct{})

	// loop
	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// if the console size has changed, handle it
				var info windows.ConsoleScreenBufferInfo
				e := windows.GetConsoleScreenBufferInfo(h, &info)
				if e != nil {
					logger.Warn("failed to get console screen buffer info", "error", e)
					continue
				}
				if info.Size.X != size.X || info.Size.Y != size.Y {
					size = info.Size

					e := s.session.WindowChange(int(info.Size.Y), int(info.Size.X))
					if e != nil {
						logger.Warn("failed to change window size", "error", e)
					}
				}
			case <-done:
				return
			}
		}
	}()

	return func() {
		// cleanup if necessary
		close(done)
	}, nil
}
