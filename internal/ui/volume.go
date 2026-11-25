package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/sebasreateguib/applemusictui/internal/models"
)

// volume is the state of changing the volume. Music takes longer to obey than
// the keys to repeat, so what is asked meanwhile is sent in one go.
type volume struct {
	sending bool // Music has not confirmed the last one sent
	target  int
}

// volumeMsg reports that sent was set, and the state after that.
type volumeMsg struct {
	sent int
	now  models.NowPlaying
}

func (a *App) changeVolume(delta int) tea.Cmd {
	if a.nowPlaying.State == "" { // not known yet
		return nil
	}
	a.volume.target = clamp(a.nowPlaying.Volume+delta, 0, 100)
	a.nowPlaying.Volume = a.volume.target // the next key starts from here
	if a.volume.sending {
		return nil // see receiveVolume
	}
	a.volume.sending = true
	return doSetVolume(a.volume.target)
}

func (a *App) receiveVolume(msg volumeMsg) []tea.Cmd {
	a.volume.sending = msg.sent != a.volume.target
	cmds := a.receiveNowPlaying(msg.now)
	if a.volume.sending {
		cmds = append(cmds, doSetVolume(a.volume.target))
	}
	return cmds
}
