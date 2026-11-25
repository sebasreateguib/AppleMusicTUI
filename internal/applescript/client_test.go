package applescript

import (
	"testing"

	"github.com/sebasreateguib/applemusictui/internal/models"
)

func TestParseTrackColumns(t *testing.T) {
	got, err := parseTrackColumns("A\x1fB || C\x1eX\x1f\x1eP\x1fQ\x1e200,5\x1f61.25")
	if err != nil {
		t.Fatal(err)
	}
	want := []models.Track{
		{Name: "A", Artist: "X", Album: "P", Duration: 200.5},
		{Name: "B || C", Album: "Q", Duration: 61.25},
	}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("got %+v", got)
	}
	for _, bad := range []string{"A\x1eB", "A\x1fB\x1eX\x1eP\x1e1"} {
		if _, err := parseTrackColumns(bad); err == nil {
			t.Errorf("no error for %q", bad)
		}
	}
}
