package applescript

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"SR-Player/internal/models"
)

func escapeAppleScriptString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	return strings.ReplaceAll(s, `"`, `\"`)
}

// run executes an AppleScript string and returns the trimmed output.
func run(script string) (string, error) {
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// runSilent executes an AppleScript and ignores output.
func runSilent(script string) error {
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

// --- Playback Controls ---

func PlayPause() error {
	return runSilent(`tell application "Music" to playpause`)
}

func NextTrack() error {
	return runSilent(`tell application "Music" to next track`)
}

func PreviousTrack() error {
	return runSilent(`tell application "Music" to previous track`)
}

func Play() error {
	return runSilent(`tell application "Music" to play`)
}

func Stop() error {
	return runSilent(`tell application "Music" to stop`)
}

// --- Volume ---

func GetVolume() (int, error) {
	out, err := run(`tell application "Music" to sound volume`)
	if err != nil {
		return 0, err
	}
	v, err := strconv.Atoi(out)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func SetVolume(v int) error {
	if v < 0 {
		v = 0
	}
	if v > 100 {
		v = 100
	}
	return runSilent(fmt.Sprintf(`tell application "Music" to set sound volume to %d`, v))
}

// --- Player State ---

func GetPlayerState() (models.PlayerState, error) {
	out, err := run(`tell application "Music" to player state as string`)
	if err != nil {
		return models.StateStopped, err
	}
	switch strings.ToLower(out) {
	case "playing":
		return models.StatePlaying, nil
	case "paused":
		return models.StatePaused, nil
	default:
		return models.StateStopped, nil
	}
}

func GetPlayerPosition() (float64, error) {
	out, err := run(`tell application "Music" to player position`)
	if err != nil {
		return 0, err
	}
	pos, err := strconv.ParseFloat(out, 64)
	if err != nil {
		return 0, err
	}
	return pos, nil
}

func SetPlayerPosition(pos float64) error {
	return runSilent(fmt.Sprintf(`tell application "Music" to set player position to %f`, pos))
}

// --- Shuffle & Repeat ---

func GetShuffle() (bool, error) {
	out, err := run(`tell application "Music" to shuffle enabled`)
	if err != nil {
		return false, err
	}
	return strings.ToLower(out) == "true", nil
}

func ToggleShuffle() error {
	return runSilent(`tell application "Music" to set shuffle enabled to not shuffle enabled`)
}

func SetShuffle(enabled bool) error {
	value := "false"
	if enabled {
		value = "true"
	}
	return runSilent(fmt.Sprintf(`tell application "Music" to set shuffle enabled to %s`, value))
}

func GetRepeatMode() (models.RepeatMode, error) {
	out, err := run(`tell application "Music" to song repeat as string`)
	if err != nil {
		return models.RepeatOff, err
	}
	switch strings.ToLower(out) {
	case "one":
		return models.RepeatOne, nil
	case "all":
		return models.RepeatAll, nil
	default:
		return models.RepeatOff, nil
	}
}

func ToggleRepeat() error {
	current, err := GetRepeatMode()
	if err != nil {
		return err
	}
	var next string
	switch current {
	case models.RepeatOff:
		next = "all"
	case models.RepeatAll:
		next = "one"
	default:
		next = "off"
	}
	return runSilent(fmt.Sprintf(`tell application "Music" to set song repeat to %s`, next))
}

// --- Current Track ---

func GetCurrentTrack() (models.Track, error) {
	script := `
tell application "Music"
	if player state is stopped then
		return ""
	end if
	set t to current track
	set trackName to name of t
	set trackArtist to artist of t
	set trackAlbum to album of t
	set trackDuration to duration of t
	set trackFormat to kind of t
	return trackName & "||" & trackArtist & "||" & trackAlbum & "||" & trackDuration & "||" & trackFormat
end tell`

	out, err := run(script)
	if err != nil || out == "" {
		return models.Track{}, err
	}

	parts := strings.Split(out, "||")
	if len(parts) < 5 {
		return models.Track{}, fmt.Errorf("unexpected output: %s", out)
	}

	duration, _ := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)

	return models.Track{
		Name:     strings.TrimSpace(parts[0]),
		Artist:   strings.TrimSpace(parts[1]),
		Album:    strings.TrimSpace(parts[2]),
		Duration: duration,
		Format:   strings.TrimSpace(parts[4]),
	}, nil
}

// GetNowPlaying returns the full playback state.
func GetNowPlaying() (models.NowPlaying, error) {
	state, _ := GetPlayerState()
	track, _ := GetCurrentTrack()
	pos, _ := GetPlayerPosition()
	shuffle, _ := GetShuffle()
	repeat, _ := GetRepeatMode()
	vol, _ := GetVolume()

	return models.NowPlaying{
		Track:          track,
		State:          state,
		Position:       pos,
		ShuffleEnabled: shuffle,
		RepeatMode:     repeat,
		Volume:         vol,
	}, nil
}

// --- Playlists ---

func GetPlaylists() ([]models.Playlist, error) {
	script := `
tell application "Music"
	set output to ""
	set allPlaylists to every playlist
	repeat with p in allPlaylists
		set pKind to special kind of p as string
		if pKind is "none" then
			try
				set pCount to count of tracks of p
				set output to output & name of p & "||" & pCount & "\n"
			end try
		end if
	end repeat
	return output
end tell`

	out, err := run(script)
	if err != nil {
		return nil, err
	}

	var playlists []models.Playlist
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "||", 2)
		if len(parts) != 2 {
			continue
		}
		count, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
		playlists = append(playlists, models.Playlist{
			Name:  strings.TrimSpace(parts[0]),
			Count: count,
		})
	}
	return playlists, nil
}

func PlayPlaylist(name string) error {
	script := fmt.Sprintf(`tell application "Music" to play playlist "%s"`, escapeAppleScriptString(name))
	return runSilent(script)
}

// GetPlaylistTracks returns tracks in the named playlist with pagination.
func GetUniqueArtists() ([]string, error) {
	script := `
tell application "Music"
	set myartists to artist of tracks of playlist "Library"
	set AppleScript's text item delimiters to "||"
	return myartists as string
end tell`

	out, err := run(script)
	if err != nil {
		return nil, err
	}

	unique := make(map[string]bool)
	var result []string
	parts := strings.Split(out, "||")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" && !unique[p] {
			unique[p] = true
			result = append(result, p)
		}
	}
	return result, nil
}

func GetUniqueAlbums() ([]string, error) {
	script := `
tell application "Music"
	set myalbums to album of tracks of playlist "Library"
	set AppleScript's text item delimiters to "||"
	return myalbums as string
end tell`

	out, err := run(script)
	if err != nil {
		return nil, err
	}

	unique := make(map[string]bool)
	var result []string
	parts := strings.Split(out, "||")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" && !unique[p] {
			unique[p] = true
			result = append(result, p)
		}
	}
	return result, nil
}

// GetFilteredTracks returns tracks paginated based on context: "playlist", "artist", "album", "library"
func GetFilteredTracks(contextType, contextValue string, offset, limit int) ([]models.Track, error) {
	startIdx := offset + 1
	endIdx := offset + limit
	escapedType := escapeAppleScriptString(contextType)
	escapedValue := escapeAppleScriptString(contextValue)

	script := fmt.Sprintf(`
tell application "Music"
	set output to ""
	
	if "%s" is "playlist" then
		set trackList to (tracks of playlist "%s")
	else if "%s" is "artist" then
		set trackList to (tracks of playlist "Library" whose artist is "%s")
	else if "%s" is "album" then
		set trackList to (tracks of playlist "Library" whose album is "%s")
	else if "%s" is "library" then
		set trackList to tracks of playlist "Library"
	end if
	
	set totalCount to count of trackList
	set startIdx to %d
	set endIdx to %d
	if startIdx > totalCount then return ""
	if endIdx > totalCount then set endIdx to totalCount
	
	repeat with i from startIdx to endIdx
		set t to item i of trackList
		set tName to name of t
		set tArtist to artist of t
		set tAlbum to album of t
		set tDuration to duration of t as string
		set output to output & tName & "||" & tArtist & "||" & tAlbum & "||" & tDuration & "\n"
	end repeat
	return output
end tell`, escapedType, escapedValue, escapedType, escapedValue, escapedType, escapedValue, escapedType, startIdx, endIdx)

	out, err := run(script)
	if err != nil || out == "" {
		return nil, err
	}

	var tracks []models.Track
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "||", 4)
		if len(parts) < 4 {
			continue
		}
		dur, _ := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
		tracks = append(tracks, models.Track{
			Name:     strings.TrimSpace(parts[0]),
			Artist:   strings.TrimSpace(parts[1]),
			Album:    strings.TrimSpace(parts[2]),
			Duration: dur,
		})
	}
	return tracks, nil
}

func GetContextTrackCount(contextType, contextValue string) (int, error) {
	escapedType := escapeAppleScriptString(contextType)
	escapedValue := escapeAppleScriptString(contextValue)

	script := fmt.Sprintf(`
tell application "Music"
	if "%s" is "playlist" then
		return (count of tracks of playlist "%s") as string
	else if "%s" is "artist" then
		return (count of (tracks of playlist "Library" whose artist is "%s")) as string
	else if "%s" is "album" then
		return (count of (tracks of playlist "Library" whose album is "%s")) as string
	else if "%s" is "library" then
		return (count of tracks of playlist "Library") as string
	end if
	return "0"
end tell`,
		escapedType,
		escapedValue,
		escapedType,
		escapedValue,
		escapedType,
		escapedValue,
		escapedType,
	)

	out, err := run(script)
	if err != nil {
		return 0, err
	}
	count, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		return 0, err
	}
	return count, nil
}

func PlayTrackInContext(contextType, contextValue string, trackIndex int) error {
	return PlayTrackInContextWithShuffle(contextType, contextValue, trackIndex, false)
}

func PlayTrackInContextWithShuffle(contextType, contextValue string, trackIndex int, shuffleAfter bool) error {
	escapedType := escapeAppleScriptString(contextType)
	escapedValue := escapeAppleScriptString(contextValue)
	shuffleValue := "false"
	if shuffleAfter {
		shuffleValue = "true"
	}
	script := fmt.Sprintf(`
tell application "Music"
	if "%s" is "playlist" then
		set thePlaylist to playlist "%s"
		set shuffle enabled to false
		play thePlaylist
		delay 0.35
		if %d > 1 then
			repeat with i from 2 to %d
				next track
				delay 0.1
			end repeat
		end if
	else if "%s" is "artist" then
		set tList to (tracks of playlist "Library" whose artist is "%s")
		play item %d of tList
	else if "%s" is "album" then
		set tList to (tracks of playlist "Library" whose album is "%s")
		play item %d of tList
	else if "%s" is "library" then
		set theLibrary to playlist "Library"
		set shuffle enabled to false
		play theLibrary
		delay 0.35
		if %d > 1 then
			repeat with i from 2 to %d
				next track
				delay 0.05
			end repeat
		end if
	end if
	set shuffle enabled to %s
end tell`, escapedType, escapedValue, trackIndex, trackIndex, escapedType, escapedValue, trackIndex, escapedType, escapedValue, trackIndex, escapedType, trackIndex, trackIndex, shuffleValue)

	return runSilent(script)
}

// PlayTrackByName plays the first track matching name in the library.
func PlayTrackByName(name string) error {
	script := fmt.Sprintf(`tell application "Music" to play track "%s"`, escapeAppleScriptString(name))
	return runSilent(script)
}

// PlayTrackInLibrary plays an exact track match from the library.
func PlayTrackInLibrary(track models.Track) error {
	script := fmt.Sprintf(`
tell application "Music"
	set tList to (tracks of playlist "Library" whose name is "%s" and artist is "%s" and album is "%s")
	if (count of tList) is 0 then error "TRACK_NOT_FOUND"
	play item 1 of tList
end tell`,
		escapeAppleScriptString(track.Name),
		escapeAppleScriptString(track.Artist),
		escapeAppleScriptString(track.Album),
	)

	return runSilent(script)
}

// FindTrackIndexInContext returns the 1-based index of the exact track inside the given context.
func FindTrackIndexInContext(contextType, contextValue string, track models.Track) (int, error) {
	escapedType := escapeAppleScriptString(contextType)
	escapedValue := escapeAppleScriptString(contextValue)
	script := fmt.Sprintf(`
tell application "Music"
	if "%s" is "playlist" then
		set trackList to (tracks of playlist "%s")
	else if "%s" is "artist" then
		set trackList to (tracks of playlist "Library" whose artist is "%s")
	else if "%s" is "album" then
		set trackList to (tracks of playlist "Library" whose album is "%s")
	else if "%s" is "library" then
		set trackList to tracks of playlist "Library"
	else
		return ""
	end if
	
	repeat with i from 1 to (count of trackList)
		set t to item i of trackList
		if (name of t is "%s") and (artist of t is "%s") and (album of t is "%s") then
			return i as string
		end if
	end repeat
	
	return ""
end tell`,
		escapedType,
		escapedValue,
		escapedType,
		escapedValue,
		escapedType,
		escapedValue,
		escapedType,
		escapeAppleScriptString(track.Name),
		escapeAppleScriptString(track.Artist),
		escapeAppleScriptString(track.Album),
	)

	out, err := run(script)
	if err != nil || out == "" {
		return 0, err
	}

	idx, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		return 0, err
	}
	return idx, nil
}

// --- Queue (current playlist tracks) ---

func GetQueueTracks() ([]models.QueueTrack, error) {
	script := `
tell application "Music"
	if player state is stopped then return ""
	try
		set t to current track
		set pl to current playlist
		set idx to index of t
		set total to count of tracks of pl
		
		set maxIdx to idx + 10
		if maxIdx > total then set maxIdx to total
		
		set output to ""
		repeat with i from idx to maxIdx
			set tr to track i of pl
			set tName to name of tr
			set tArtist to artist of tr
			set isCurrent to "0"
			if i is idx then set isCurrent to "1"
			set output to output & i & "||" & tName & "||" & tArtist & "||" & isCurrent & "\n"
		end repeat
		return output
	on error
		return ""
	end try
end tell`

	out, err := run(script)
	if err != nil || out == "" {
		return nil, err
	}

	var tracks []models.QueueTrack
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "||", 4)
		if len(parts) < 4 {
			continue
		}
		idx, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
		tracks = append(tracks, models.QueueTrack{
			Index: idx,
			Track: models.Track{
				Name:   strings.TrimSpace(parts[1]),
				Artist: strings.TrimSpace(parts[2]),
			},
			IsCurrent: strings.TrimSpace(parts[3]) == "1",
		})
	}
	return tracks, nil
}

// GetQueueTracksForContext returns the current and upcoming tracks for a known playback context.
func GetQueueTracksForContext(contextType, contextValue string, currentIndex, ahead int) ([]models.QueueTrack, error) {
	if currentIndex < 1 {
		return nil, nil
	}
	if ahead < 0 {
		ahead = 0
	}

	escapedType := escapeAppleScriptString(contextType)
	escapedValue := escapeAppleScriptString(contextValue)
	script := fmt.Sprintf(`
tell application "Music"
	if "%s" is "playlist" then
		set trackList to (tracks of playlist "%s")
	else if "%s" is "artist" then
		set trackList to (tracks of playlist "Library" whose artist is "%s")
	else if "%s" is "album" then
		set trackList to (tracks of playlist "Library" whose album is "%s")
	else if "%s" is "library" then
		set trackList to tracks of playlist "Library"
	else
		return ""
	end if
	
	set total to count of trackList
	if %d > total then return ""
	
	set maxIdx to %d + %d
	if maxIdx > total then set maxIdx to total
	
	set output to ""
	repeat with i from %d to maxIdx
		set tr to item i of trackList
		set tName to name of tr
		set tArtist to artist of tr
		set isCurrent to "0"
		if i is %d then set isCurrent to "1"
		set output to output & i & "||" & tName & "||" & tArtist & "||" & isCurrent & "\n"
	end repeat
	return output
end tell`,
		escapedType,
		escapedValue,
		escapedType,
		escapedValue,
		escapedType,
		escapedValue,
		escapedType,
		currentIndex,
		currentIndex,
		ahead,
		currentIndex,
		currentIndex,
	)

	out, err := run(script)
	if err != nil || out == "" {
		return nil, err
	}

	var tracks []models.QueueTrack
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "||", 4)
		if len(parts) < 4 {
			continue
		}
		idx, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
		tracks = append(tracks, models.QueueTrack{
			Index: idx,
			Track: models.Track{
				Name:   strings.TrimSpace(parts[1]),
				Artist: strings.TrimSpace(parts[2]),
			},
			IsCurrent: strings.TrimSpace(parts[3]) == "1",
		})
	}
	return tracks, nil
}

// --- Search ---

func SearchLibrary(query string) ([]models.Track, error) {
	escapedQuery := escapeAppleScriptString(query)
	script := fmt.Sprintf(`
tell application "Music"
	set results to search playlist "Library" for "%s"
	set output to ""
	set i to 0
	repeat with t in results
		set output to output & name of t & "||" & artist of t & "||" & album of t & "\n"
		set i to i + 1
		if i >= 30 then exit repeat
	end repeat
	return output
end tell`, escapedQuery)

	out, err := run(script)
	if err != nil || out == "" {
		return nil, err
	}

	var tracks []models.Track
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "||", 3)
		if len(parts) < 3 {
			continue
		}
		tracks = append(tracks, models.Track{
			Name:   strings.TrimSpace(parts[0]),
			Artist: strings.TrimSpace(parts[1]),
			Album:  strings.TrimSpace(parts[2]),
		})
	}
	return tracks, nil
}

// --- Artwork ---

// GetArtworkPath extracts the current track's artwork to a unique JPEG file and
// returns its path. Returns "" if no artwork is available.
func GetArtworkPath() string {
	// Clean up any old artwork files to prevent filling up /tmp
	oldFiles, _ := filepath.Glob("/tmp/sr-player-art-*.jpg")
	for _, f := range oldFiles {
		os.Remove(f)
	}

	// Check if Music is running and playing
	stateOut, err := run(`tell application "Music" to player state as string`)
	if err != nil || strings.ToLower(stateOut) == "stopped" {
		return ""
	}

	// Generate a guaranteed unique path for cache-busting the UI
	artworkCachePath := fmt.Sprintf("/tmp/sr-player-art-%d.jpg", time.Now().UnixNano())

	script := fmt.Sprintf(`
tell application "Music"
	try
		set t to current track
		set artList to artworks of t
		if (count of artList) is 0 then return ""
		set art to item 1 of artList
		set artData to raw data of art
		set filePath to "%s"
		set fileRef to open for access POSIX file filePath with write permission
		set eof fileRef to 0
		write artData to fileRef
		close access fileRef
		return filePath
	on error
		return ""
	end try
end tell`, artworkCachePath)

	out, err := run(script)
	if err != nil || out == "" {
		return ""
	}

	// Verify the file exists and has content
	info, err := os.Stat(artworkCachePath)
	if err != nil || info.Size() == 0 {
		return ""
	}

	return artworkCachePath
}
