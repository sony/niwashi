//go:build !windows

package ssh

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sony/niwashi/internal/logger"
	"golang.org/x/term"
)

func handleShellResize(s *sshSession) (func(), error) {
	// resize event
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)
	go func() {
		for range sigCh {
			width, height, e := term.GetSize(int(os.Stdout.Fd()))
			if e != nil {
				logger.Warn("failed to get terminal size", "error", e)
				continue
			}
			e = s.session.WindowChange(height, width)
			if e != nil {
				logger.Warn("failed to change window size", "error", e)
			}
		}
	}()

	return func() {
		signal.Stop(sigCh)
		close(sigCh)
	}, nil
}
