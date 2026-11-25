# 🎵 Apple Music TUI

> **⚠️ Work in Progress** — This project is actively under development. Features may be incomplete, unstable, or subject to change.

A terminal-based Apple Music player for macOS, built with Go. Browse your library and control playback from the command line, with a panel layout, real album art and a visualizer that follows the actual sound.

![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)
![Platform](https://img.shields.io/badge/Platform-macOS-lightgrey?style=flat&logo=apple)
![Status](https://img.shields.io/badge/Status-In%20Development-orange?style=flat)

<div align="center">
  <img src="assets/demo.gif" width=800 />
  <br/>
</div>

## ✨ Features

**Library**

- 🪟 **Panel layout** — Playlists, Artists, Albums, Tracks and Up Next on one screen, with a live preview of whatever is selected
- 📚 **Full library browsing** — every song, playlist, artist and album, without paging
- 🔍 **Search** — find songs by title, artist or album (up to 30 results)

**Playback**

- 🖱️ **Seek with the mouse** — click or drag the progress bar to restart the song or jump to any moment
- 🔀 **Shuffle & repeat** — a shuffle order of its own, so Up Next shows what really comes next; repeat off / all / one
- 📜 **Up Next** — up to 30 upcoming tracks of the list you are playing
- 🔊 **Volume control** — in steps of 5 %

**Looks**

- 🎧 **Now Playing** — cover, title, artist, album and an audio format badge
- 🖼️ **Album art** — drawn with Unicode blocks in any terminal, or as the real cover in Ghostty and Kitty (press `i`; the choice is remembered)
- 📊 **Visualizer** — bars driven by the real audio output through [cava](https://github.com/karlstav/cava), at 60 fps
- ⚡ **Light on the CPU** — panels are only rebuilt when something in them changes
- 📐 **Adapts to the terminal** — from 60×16 cells up

<br/>

## 📋 Requirements

- **macOS** with the **Music** app (controlled through AppleScript)
- **Go 1.24+** to build
- A terminal with **true color** support (Ghostty, Kitty, iTerm2, Warp, etc.)
- **cava 1.0+** for the visualizer — optional, needs macOS 14.2+

```bash
brew install cava
```

### Permissions

macOS asks for these on first use. Both can be changed in *System Settings → Privacy & Security*:

| Permission | Where | Needed for |
|------------|-------|------------|
| Control Music | *Automation* | Everything |
| Record system audio | *Screen & System Audio Recording* | The visualizer |

Grant them to your terminal. cava reads the system output directly, with no virtual audio device, and your own cava configuration is not used or changed.

<br/>

## 🚀 Getting Started

```bash
git clone https://github.com/sebasreateguib/AppleMusicTUI.git
cd AppleMusicTUI

go build -o applemusictui .
./applemusictui
```

Or without keeping a binary:

```bash
go run .
```

If Music is closed, the player shows its logo while it opens Music in the background, with no window. If it is already open, you go straight in.

<br/>

## ⌨️ Keybindings

Press `?` in the player to see them all.

**Panels**

| Key | Action |
|-----|--------|
| `1` – `5` | Focus Playlists / Artists / Albums / Tracks / Up Next |
| `←` `→` / `h` `l` / `Tab` `Shift+Tab` | Previous / next panel |
| `Enter` | Open the selection / play the track |
| `Esc` | Back to the list on the left |

**Lists**

| Key | Action |
|-----|--------|
| `↑` `↓` / `j` `k` | Move |
| `PgUp` `PgDn` / `Ctrl+U` `Ctrl+D` | Move a page |
| `g` `G` / `Home` `End` | First / last |
| `c` | Go to the track that is playing |

**Playback**

| Key | Action |
|-----|--------|
| `Space` | Play / pause |
| `n` `p` | Next / previous track |
| `+` `-` | Volume |
| `s` | Shuffle |
| `r` | Repeat: off / all / one |

**Other**

| Key | Action |
|-----|--------|
| `/` | Search the library |
| `i` | Album art: pixelated / photo (Ghostty and Kitty only) |
| `R` | Refresh the library |
| `?` | Help |
| `q` / `Ctrl+C` | Quit |

### 🖱️ Mouse

| Action | Effect |
|--------|--------|
| Click the progress bar | Jump to that moment (the left end restarts the song) |
| Drag the progress bar | Preview the time; the song jumps when you release |
| Wheel | Move in the focused list |

To select text with the mouse, hold `Shift` while dragging (`Option` in iTerm2).

<br/>

## 📐 Terminal size

| Size | What is shown |
|------|---------------|
| 100 columns or more | Everything: lists, Tracks, Now Playing and Up Next |
| 60 – 99 columns | Now Playing is hidden, and Up Next (`5`) replaces Tracks while focused |
| 24 rows or more | The visualizer |
| Less than 60×16 | A message asking for more room |

<br/>

## 🩺 Troubleshooting

| Problem | Solution |
|---------|----------|
| `cava not found` in the visualizer | `brew install cava` |
| `No audio captured` in the visualizer | Allow your terminal under *Screen & System Audio Recording*, then restart it |
| `Photo artwork needs Ghostty or Kitty` | Photo mode needs one of those terminals, and not inside tmux or another multiplexer |
| `Music did not answer` | Check that Music is open and that *Automation* is allowed, then press `R` |
| Up Next says `Shuffled by Music` | Music does not tell apps its own shuffle order. Play a track from the player, or toggle shuffle with `s` |

<br/>

## 🗂️ Project Structure

```
.
├── main.go                  # Entry point
├── splash.go                # Logo at startup, opens Music
├── assets/
│   ├── am-ascii.txt         # Apple ASCII art
│   └── demo.gif
└── internal/
    ├── applescript/         # macOS AppleScript bridge (playback, library queries)
    ├── artwork/             # Real album art (Kitty graphics protocol)
    ├── cava/                # cava child process (real audio levels)
    ├── config/              # Saved preferences
    ├── models/              # Shared data types (Track, Playlist, NowPlaying, etc.)
    └── ui/                  # TUI layout and rendering
        ├── app.go           # Main Bubble Tea model (Update loop, View)
        ├── keys.go          # Keybindings
        ├── commands.go      # Calls to Music
        ├── playback.go      # Playback context, shuffle order and Up Next
        ├── seek.go          # Mouse: seeking and the wheel
        ├── layout.go        # Panel sizing
        ├── statusbar.go     # Player bar and bottom line
        ├── viewcache.go     # Redraws only what changed
        ├── theme/           # Colors
        └── panels/          # Frame, lists, now playing, visualizer, popups
```

Preferences are stored in `~/Library/Application Support/applemusictui/config.json`.

Run the tests with:

```bash
go test ./...
```

<br/>

## 🛠️ Built With

| Project | Purpose |
|---------|---------|
| [Bubble Tea](https://github.com/charmbracelet/bubbletea) | TUI framework (Elm-like architecture) |
| [Lipgloss](https://github.com/charmbracelet/lipgloss) | Terminal styling |
| [Bubbles](https://github.com/charmbracelet/bubbles) | Text input |
| [x/ansi](https://github.com/charmbracelet/x) | Measuring and cutting styled text |
| [cava](https://github.com/karlstav/cava) | Audio analysis for the visualizer |

<br/>

## 🚧 Next up:

- [ ] Queue editing (reorder / remove tracks)
- [ ] Lyrics view
- [ ] Playlist creation / management

<br/>

## 📄 License

MIT — see `LICENSE` for details.
