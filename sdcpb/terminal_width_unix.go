//go:build !windows

package sdcpb

import (
	"os"

	"golang.org/x/sys/unix"
)

func detectTerminalWidth() int {
	for _, fd := range []uintptr{os.Stdout.Fd(), os.Stderr.Fd(), os.Stdin.Fd()} {
		ws, err := unix.IoctlGetWinsize(int(fd), unix.TIOCGWINSZ)
		if err == nil && ws != nil && ws.Col > 0 {
			return int(ws.Col)
		}
	}
	return 0
}
