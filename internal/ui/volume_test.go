package ui

import (
	"testing"

	"github.com/sebasreateguib/applemusictui/internal/models"
)

func TestVolumeKeysAddUp(t *testing.T) {
	np := models.NowPlaying{State: models.StatePlaying, Volume: 50}
	a := send(t, newTestApp(t), nowPlayingMsg(np))

	sent := 0
	for range 3 {
		if a.changeVolume(-5) != nil {
			sent++
		}
	}
	if a.nowPlaying.Volume != 35 || sent != 1 {
		t.Fatalf("volume %d after three keys, %d sent", a.nowPlaying.Volume, sent)
	}

	// Reports from before Music obeyed do not move it back.
	a = send(t, a, polledMsg(np))
	if a.nowPlaying.Volume != 35 {
		t.Fatalf("an old report moved it to %d", a.nowPlaying.Volume)
	}

	// The first one is confirmed: the rest has to be sent.
	np.Volume = 45
	a = send(t, a, volumeMsg{45, np})
	if a.nowPlaying.Volume != 35 || !a.volume.sending {
		t.Fatalf("volume %d, sending %v", a.nowPlaying.Volume, a.volume.sending)
	}
	np.Volume = 35
	a = send(t, a, volumeMsg{35, np})
	if a.volume.sending {
		t.Fatal("still sending")
	}

	// Music is listened to again.
	np.Volume = 80
	if a = send(t, a, polledMsg(np)); a.nowPlaying.Volume != 80 {
		t.Fatalf("volume %d", a.nowPlaying.Volume)
	}
	a.changeVolume(50)
	if a.nowPlaying.Volume != 100 {
		t.Fatalf("volume %d above the maximum", a.nowPlaying.Volume)
	}
}

func TestVolumeUnknown(t *testing.T) {
	a := newTestApp(t)
	if a.changeVolume(5) != nil {
		t.Fatal("set from an unknown volume")
	}
}
