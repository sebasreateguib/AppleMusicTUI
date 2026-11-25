package applescript

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sebasreateguib/applemusictui/internal/models"
)

// Every launch of osascript makes macOS check permissions and signatures again,
// which costs far more than the question. The playback state is asked for all
// the time, so a single process stays open and answers each byte written to its
// input. It ends by itself when that input is closed, even if the app is killed.
//
// It is not JavaScript, which leaks about 50 KB per answer. An error in the
// player state ends it, and the next question starts another one.
const monitorScript = `
use framework "Foundation"
use scripting additions
set theInput to current application's NSFileHandle's fileHandleWithStandardInput()
set theOutput to current application's NSFileHandle's fileHandleWithStandardOutput()
set text item delimiters to (character id 31)
repeat
	if (theInput's availableData()'s |length|()) as integer is 0 then exit repeat
	set cols to {"", "", "", 0, ""}
	set shuffleOn to false
	set repeatMode to "off"
	set vol to 0
	set pos to 0
	tell application "Music"
		set theState to player state as string
		if theState is not "stopped" then
			try
				set t to current track
				set cols to {name of t, artist of t, album of t, duration of t, kind of t}
			end try
		end if
		try
			set shuffleOn to shuffle enabled
		end try
		try
			set repeatMode to song repeat as string
		end try
		try
			set vol to sound volume
		end try
		try
			set pos to player position
		end try
	end tell
	set answer to ({theState} & cols & {shuffleOn, repeatMode, vol, pos}) as string
	theOutput's writeData:((current application's NSString's stringWithString:(answer & (character id 30)))'s dataUsingEncoding:4)
end repeat`

var errMonitorClosed = errors.New("monitor closed")

// monitor is a process that answers one question, as many times as it is asked.
type monitor struct {
	command []string

	mu     sync.Mutex // one question at a time
	stdin  io.WriteCloser
	stdout *bufio.Reader

	procMu sync.Mutex // not mu, which is held while waiting for an answer
	cmd    *exec.Cmd
	closed bool
}

var playerMonitor = &monitor{command: []string{"osascript", "-e", monitorScript}}

// Close ends the process that GetNowPlaying keeps open.
func Close() { playerMonitor.close() }

// ask returns the answer of the process, which is started if needed.
func (m *monitor) ask() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var err error
	for range 2 { // the process may have died since the last question
		var answer string
		if answer, err = m.try(); err == nil {
			return answer, nil
		}
		m.stop()
		if errors.Is(err, errMonitorClosed) {
			break
		}
	}
	return "", err
}

func (m *monitor) try() (string, error) {
	if err := m.start(); err != nil {
		return "", err
	}
	if _, err := m.stdin.Write([]byte{'\n'}); err != nil {
		return "", err
	}
	answer, err := m.stdout.ReadString('\x1e')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(answer, "\x1e"), nil
}

// start launches the process, unless it is running.
func (m *monitor) start() error {
	m.procMu.Lock()
	defer m.procMu.Unlock()
	if m.closed {
		return errMonitorClosed
	}
	if m.cmd != nil {
		return nil
	}
	cmd := exec.Command(m.command[0], m.command[1:]...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	m.cmd, m.stdin, m.stdout = cmd, stdin, bufio.NewReader(stdout)
	return nil
}

// stop ends the process, if any.
func (m *monitor) stop() {
	m.procMu.Lock()
	defer m.procMu.Unlock()
	if m.cmd == nil {
		return
	}
	m.cmd.Process.Kill()
	m.cmd.Wait()
	m.cmd = nil
}

func (m *monitor) close() {
	m.procMu.Lock()
	m.closed = true
	m.procMu.Unlock()
	m.stop()
}

// GetNowPlaying returns the full playback state.
func GetNowPlaying() (models.NowPlaying, error) {
	stopped := models.NowPlaying{State: models.StateStopped, RepeatMode: models.RepeatOff}
	answer, err := playerMonitor.ask()
	if err != nil {
		return stopped, err
	}
	np, err := parseNowPlaying(answer)
	if err != nil {
		return stopped, err
	}
	np.PositionAt = time.Now() // the position is the last thing it reads
	return np, nil
}

func parseNowPlaying(answer string) (models.NowPlaying, error) {
	f := strings.Split(answer, "\x1f")
	if len(f) != 10 {
		return models.NowPlaying{}, fmt.Errorf("unexpected reply from Music: %d fields", len(f))
	}
	volume, _ := strconv.Atoi(f[8])
	return models.NowPlaying{
		State: parseState(f[0]),
		Track: models.Track{
			Name:     strings.TrimSpace(f[1]),
			Artist:   strings.TrimSpace(f[2]),
			Album:    strings.TrimSpace(f[3]),
			Duration: parseDecimal(f[4]),
			Format:   strings.TrimSpace(f[5]),
		},
		ShuffleEnabled: f[6] == "true",
		RepeatMode:     parseRepeat(f[7]),
		Volume:         volume,
		Position:       parseDecimal(f[9]),
	}, nil
}

func parseState(s string) models.PlayerState {
	switch strings.ToLower(s) {
	case "playing":
		return models.StatePlaying
	case "paused":
		return models.StatePaused
	default:
		return models.StateStopped
	}
}

func parseRepeat(s string) models.RepeatMode {
	switch strings.ToLower(s) {
	case "one":
		return models.RepeatOne
	case "all":
		return models.RepeatAll
	default:
		return models.RepeatOff
	}
}

// parseDecimal reads a number written the way the user's region does.
func parseDecimal(s string) float64 {
	v, _ := strconv.ParseFloat(strings.Replace(strings.TrimSpace(s), ",", ".", 1), 64)
	return v
}
