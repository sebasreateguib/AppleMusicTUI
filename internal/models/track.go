package models

import "time"

// PlayerState represents the current state of the Music app player.
type PlayerState string

const (
	StatePlaying PlayerState = "playing"
	StatePaused  PlayerState = "paused"
	StateStopped PlayerState = "stopped"
)

// RepeatMode represents the repeat setting.
type RepeatMode string

const (
	RepeatOff RepeatMode = "off"
	RepeatOne RepeatMode = "one"
	RepeatAll RepeatMode = "all"
)

// Track holds metadata for a single music track.
type Track struct {
	Name     string
	Artist   string
	Album    string
	Duration float64 // seconds
	Format   string
}

// Same reports whether both are the same track. Details such as the format are
// not always known.
func (t Track) Same(o Track) bool {
	return t.Name == o.Name && t.Artist == o.Artist && t.Album == o.Album
}

// Playlist holds metadata for a playlist.
type Playlist struct {
	Name  string
	Count int
}

// NowPlaying holds the current playback state.
type NowPlaying struct {
	Track          Track
	State          PlayerState
	Position       float64   // current position in seconds
	PositionAt     time.Time // when it was read, zero if unknown
	ShuffleEnabled bool
	RepeatMode     RepeatMode
	Volume         int    // 0-100
	ArtworkPath    string // path to temp JPEG artwork file, "" if none
}

// QueueTrack is a track entry in the current queue.
type QueueTrack struct {
	Index     int
	Track     Track
	IsCurrent bool
}
