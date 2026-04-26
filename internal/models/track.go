package models

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

// Playlist holds metadata for a playlist.
type Playlist struct {
	Name  string
	Count int
}

// NowPlaying holds the current playback state.
type NowPlaying struct {
	Track          Track
	State          PlayerState
	Position       float64 // current position in seconds
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
