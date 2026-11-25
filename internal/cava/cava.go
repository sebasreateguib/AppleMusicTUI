// Package cava runs the cava audio visualizer as a child process and exposes
// the bar levels it computes from the real system audio output.
package cava

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	framerate      = 60
	maxBars        = 200
	resizeDebounce = 200 * time.Millisecond
)

// ErrNotInstalled is reported when the cava binary is not in PATH.
var ErrNotInstalled = errors.New("cava not found - run: brew install cava")

// Event is either a frame of bar levels (0-255, low to high frequency) or the
// error that stopped cava.
type Event struct {
	Levels []byte
	Err    error
}

// Visualizer supervises a single cava process.
type Visualizer struct {
	events    chan Event
	resize    chan struct{}
	quit      chan struct{}
	done      chan struct{}
	closeOnce sync.Once
	bars      atomic.Int32
	active    atomic.Bool
}

// New returns a Visualizer. cava is not launched until the first Resize.
func New() *Visualizer {
	v := &Visualizer{
		events: make(chan Event, 1),
		resize: make(chan struct{}, 1),
		quit:   make(chan struct{}),
		done:   make(chan struct{}),
	}
	v.active.Store(true)
	go v.supervise()
	return v
}

// Events delivers the most recent frame; frames not consumed in time are dropped.
func (v *Visualizer) Events() <-chan Event { return v.events }

// Resize sets the number of bars, (re)starting cava when it changes.
func (v *Visualizer) Resize(bars int) {
	bars = min(bars, maxBars)
	if bars < 1 || v.bars.Swap(int32(bars)) == int32(bars) {
		return
	}
	select {
	case v.resize <- struct{}{}:
	default:
	}
}

// SetActive toggles frame delivery, so hidden visualizers trigger no redraws.
func (v *Visualizer) SetActive(active bool) { v.active.Store(active) }

// Close stops cava and waits for it to exit.
func (v *Visualizer) Close() {
	v.closeOnce.Do(func() { close(v.quit) })
	<-v.done
}

func (v *Visualizer) supervise() {
	defer close(v.done)
	stop := func() {}
	defer func() { stop() }()

	current := 0
	for {
		select {
		case <-v.quit:
			return
		case <-v.resize:
		}
		if current != 0 {
			// Dragging a window edge emits many resizes; restart once.
			select {
			case <-v.quit:
				return
			case <-time.After(resizeDebounce):
			}
		}
		bars := int(v.bars.Load())
		if bars == current {
			continue
		}
		stop()
		current = bars
		stop = v.start(bars)
	}
}

// start launches cava in the background and returns a func that stops it and
// waits for it to exit.
func (v *Visualizer) start(bars int) func() {
	ctx, cancel := context.WithCancel(context.Background())
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		if err := v.run(ctx, bars); err != nil && ctx.Err() == nil {
			v.emit(Event{Err: err})
		}
	}()
	return func() {
		cancel()
		<-finished
	}
}

func (v *Visualizer) run(ctx context.Context, bars int) error {
	bin, err := exec.LookPath("cava")
	if err != nil {
		return ErrNotInstalled
	}
	cfg, err := writeConfig(bars)
	if err != nil {
		return err
	}
	defer os.Remove(cfg)

	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, bin, "-p", cfg)
	cmd.Stderr = &stderr
	cmd.Cancel = func() error { return cmd.Process.Signal(syscall.SIGTERM) }
	cmd.WaitDelay = time.Second
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	r := bufio.NewReader(stdout)
	wasSilent := false
	for {
		levels := make([]byte, bars)
		if _, err := io.ReadFull(r, levels); err != nil {
			break
		}
		if !v.active.Load() {
			wasSilent = false
			continue
		}
		// Repeated silence carries no information; skip the redraw.
		silent := IsSilent(levels)
		if silent && wasSilent {
			continue
		}
		wasSilent = silent
		v.emit(Event{Levels: levels})
	}

	err = cmd.Wait()
	if msg := lastLine(stderr.String()); msg != "" {
		return fmt.Errorf("cava: %s", msg)
	}
	if err != nil {
		return fmt.Errorf("cava stopped: %w", err)
	}
	return errors.New("cava stopped")
}

// emit publishes e, replacing a pending event nobody has consumed yet.
func (v *Visualizer) emit(e Event) {
	select {
	case <-v.events:
	default:
	}
	select {
	case v.events <- e:
	default:
	}
}

// IsSilent reports whether every bar is at zero.
func IsSilent(levels []byte) bool {
	for _, l := range levels {
		if l != 0 {
			return false
		}
	}
	return true
}

// writeConfig creates a private config so the user's own cava config is
// neither used nor modified. source=tap captures the system output directly
// (macOS 14.2+), with no virtual audio device.
func writeConfig(bars int) (string, error) {
	f, err := os.CreateTemp("", "sr-player-cava-*.conf")
	if err != nil {
		return "", err
	}
	_, err = fmt.Fprintf(f, `[general]
framerate = %d
bars = %d
[input]
method = coreaudio
source = tap
[output]
method = raw
raw_target = /dev/stdout
data_format = binary
bit_format = 8bit
channels = mono
`, framerate, bars)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}

func lastLine(s string) string {
	s = strings.TrimSpace(s)
	return strings.TrimSpace(s[strings.LastIndexByte(s, '\n')+1:])
}
