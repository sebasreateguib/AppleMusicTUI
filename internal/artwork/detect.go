package artwork

import (
	"bytes"
	"os"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

// queryTimeout covers terminals that never answer.
const queryTimeout = 500 * time.Millisecond

// Supported asks the terminal whether it can show real images. It reads from
// stdin, so it must run before Bubble Tea starts.
//
// Environment variables are not enough: multiplexers inherit TERM_PROGRAM from
// the terminal they run in, or imitate it, without drawing images themselves.
func Supported() bool {
	// Apple's Terminal prints the queries it does not understand.
	if os.Getenv("TERM_PROGRAM") == "Apple_Terminal" {
		return false
	}

	// Other terminals have the graphics protocol but not its placeholders, so
	// the name matters. The trailing request (DA1) is answered by every
	// terminal and marks the end of the reply.
	if !hasPlaceholders(query("\x1b[>0q\x1b[c")) {
		return false
	}
	return bytes.Contains(query("\x1b_Gi=31,s=1,v=1,a=q,t=d,f=24;AAAA\x1b\\\x1b[c"), []byte("\x1b_Gi=31;OK"))
}

// hasPlaceholders reads the terminal name from an XTVERSION reply, such as
// "\x1bP>|ghostty 1.4.0\x1b\\" or "\x1bP>|kitty(0.40.0)\x1b\\".
func hasPlaceholders(reply []byte) bool {
	_, name, ok := strings.Cut(string(reply), "\x1bP>|")
	if !ok {
		return false
	}
	name = strings.ToLower(name)
	return strings.HasPrefix(name, "ghostty") || strings.HasPrefix(name, "kitty")
}

// query sends seq, which must end with a DA1 request, and returns the reply.
func query(seq string) []byte {
	fd := int(os.Stdin.Fd())
	saved, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	if err != nil {
		return nil // not a terminal
	}
	if _, err := unix.IoctlGetTermios(int(os.Stdout.Fd()), unix.TIOCGETA); err != nil {
		return nil
	}

	raw := *saved
	raw.Lflag &^= unix.ECHO | unix.ICANON
	raw.Cc[unix.VMIN] = 0
	raw.Cc[unix.VTIME] = 1 // reads return after 100 ms without input
	if err := unix.IoctlSetTermios(fd, unix.TIOCSETA, &raw); err != nil {
		return nil
	}
	defer unix.IoctlSetTermios(fd, unix.TIOCSETA, saved)

	if err := write(seq); err != nil {
		return nil
	}

	var reply []byte
	buf := make([]byte, 256)
	for deadline := time.Now().Add(queryTimeout); time.Now().Before(deadline); {
		n, err := unix.Read(fd, buf)
		if err != nil && err != unix.EINTR {
			break
		}
		if n > 0 {
			reply = append(reply, buf[:n]...)
		}
		if answered(reply) {
			break
		}
	}
	return reply
}

// answered reports whether reply contains a whole DA1 answer: "\x1b[?...c".
func answered(reply []byte) bool {
	_, rest, ok := bytes.Cut(reply, []byte("\x1b[?"))
	return ok && bytes.IndexByte(rest, 'c') >= 0
}
