package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"

	"github.com/sebasreateguib/applemusictui/assets"
)

const (
	appName = "AppleMusicTUI"

	// logoLineDelay draws the logo line by line.
	logoLineDelay = 40 * time.Millisecond

	// musicTimeout is how long Music may take to open its library.
	musicTimeout = 20 * time.Second

	red   = "\033[38;2;250;36;60m"
	bold  = "\033[1m"
	dim   = "\033[2m"
	reset = "\033[0m"
)

func musicRunning() bool {
	return exec.Command("pgrep", "-x", "Music").Run() == nil
}

// ensureMusic opens Music if it is closed, showing the logo meanwhile.
func ensureMusic() error {
	if musicRunning() {
		return nil
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// The alternate screen, so that what the terminal had comes back.
	fmt.Print("\033[?1049h\033[?25l\033[2J")
	defer fmt.Print("\033[?25h\033[?1049l")
	printSplash("Starting Apple Music…")

	// -g: without taking the focus; -j: hidden, so no window shows up.
	if err := exec.CommandContext(ctx, "open", "-g", "-j", "-a", "Music").Run(); err != nil {
		if ctx.Err() != nil {
			return context.Canceled
		}
		return errors.New("could not open Apple Music")
	}
	return waitForMusic(ctx)
}

// waitForMusic waits until Music has its library, because the player asks for
// it only once, when it starts.
func waitForMusic(ctx context.Context) error {
	wait, cancel := context.WithTimeout(ctx, musicTimeout)
	defer cancel()
	for {
		// Asking a closed Music would open it, with its window.
		if musicRunning() && libraryReady(wait) {
			return nil
		}
		select {
		case <-wait.Done():
			if ctx.Err() != nil {
				return context.Canceled
			}
			if !musicRunning() {
				return errors.New("Apple Music did not start")
			}
			return nil // slow, or waiting for a permission: the player can retry
		case <-time.After(200 * time.Millisecond):
		}
	}
}

func libraryReady(ctx context.Context) bool {
	out, err := exec.CommandContext(ctx, "osascript", "-e", `tell application "Music" to (count of playlists) > 0`).Output()
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

// printSplash draws the logo, the name and status in the middle of the screen.
func printSplash(status string) {
	lines := logoLines()
	logoW := 0
	for _, l := range lines {
		logoW = max(logoW, len(l))
	}

	cols, rows := 80, 24
	if ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ); err == nil && ws.Col > 0 && ws.Row > 0 {
		cols, rows = int(ws.Col), int(ws.Row)
	}
	row := max(1, (rows-len(lines)-4)/2+1)
	// put writes text, width cells wide, centered in the next row.
	put := func(style, text string, width int) {
		if row <= rows {
			fmt.Printf("\033[%d;%dH%s%s%s", row, max(1, (cols-width)/2+1), style, text, reset)
		}
		row++
	}

	for _, l := range lines {
		put(red, l, logoW)
		time.Sleep(logoLineDelay)
	}
	row++
	put(bold, appName, len(appName))
	row++
	put(dim, status, len([]rune(status)))
}

// logoLines returns the logo without the margin on its left.
func logoLines() []string {
	lines := strings.Split(strings.TrimRight(assets.Logo, "\n"), "\n")
	margin := len(assets.Logo)
	for _, l := range lines {
		margin = min(margin, len(l)-len(strings.TrimLeft(l, " ")))
	}
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i][margin:], " ")
	}
	return lines
}
