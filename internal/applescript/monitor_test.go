package applescript

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sebasreateguib/applemusictui/internal/models"
)

func TestParseNowPlaying(t *testing.T) {
	fields := []string{"playing", " Song ", "A || B", "Al\nbum", "200,5", "AAC", "true", "one", "70", "12.25"}
	got, err := parseNowPlaying(strings.Join(fields, "\x1f"))
	if err != nil {
		t.Fatal(err)
	}
	want := models.NowPlaying{
		Track:          models.Track{Name: "Song", Artist: "A || B", Album: "Al\nbum", Duration: 200.5, Format: "AAC"},
		State:          models.StatePlaying,
		Position:       12.25,
		ShuffleEnabled: true,
		RepeatMode:     models.RepeatOne,
		Volume:         70,
	}
	if got != want {
		t.Fatalf("got %+v", got)
	}
	if _, err := parseNowPlaying("playing"); err == nil {
		t.Error("no error for a short reply")
	}
}

// shell returns a monitor that runs script, which can count its launches in $1.
func shell(t *testing.T, script string) (m *monitor, launches func() int) {
	log := filepath.Join(t.TempDir(), "log")
	m = &monitor{command: []string{"sh", "-c", `echo x >> "$1"; ` + script, "sh", log}}
	t.Cleanup(m.close)
	return m, func() int {
		data, _ := os.ReadFile(log)
		return strings.Count(string(data), "x")
	}
}

const answering = `n=0; while read -r _; do n=$((n+1)); printf 'answer %d\036' $n; done`

func TestMonitorKeepsOneProcess(t *testing.T) {
	m, launches := shell(t, answering)
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := m.ask(); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if got, err := m.ask(); got != "answer 21" || err != nil {
		t.Errorf("got %q, %v", got, err)
	}
	if n := launches(); n != 1 {
		t.Errorf("%d launches", n)
	}
}

func TestMonitorRestarts(t *testing.T) {
	m, launches := shell(t, `read -r _; printf 'ok\036'`) // dies after answering
	for range 3 {
		if got, err := m.ask(); got != "ok" || err != nil {
			t.Fatalf("got %q, %v", got, err)
		}
	}
	if n := launches(); n != 3 {
		t.Errorf("%d launches", n)
	}

	m, launches = shell(t, `exit 1`)
	if _, err := m.ask(); err == nil {
		t.Error("no error")
	}
	if n := launches(); n != 2 {
		t.Errorf("%d launches, want a single retry", n)
	}
}

func TestMonitorCloseWhileWaiting(t *testing.T) {
	m, launches := shell(t, `exec sleep 60`) // never answers
	done := make(chan error)
	go func() {
		_, err := m.ask()
		done <- err
	}()
	for launches() == 0 {
		time.Sleep(time.Millisecond)
	}
	m.close()
	select {
	case err := <-done:
		if err == nil {
			t.Error("no error")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("close did not interrupt the question")
	}
	if n := launches(); n != 1 {
		t.Errorf("%d launches: started again after close", n)
	}
}
