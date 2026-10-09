package journal

import (
	"os"
	"testing"
	"time"
)

func TestParseBoots(t *testing.T) {
	data, err := os.ReadFile("../../testdata/list-boots.json")
	if err != nil {
		t.Fatal(err)
	}
	boots, err := parseBoots(data)
	if err != nil {
		t.Fatal(err)
	}

	if len(boots) != 3 {
		t.Fatalf("got %d boots, want 3", len(boots))
	}
	// oldest first, as journalctl prints them
	for i, want := range []int{-2, -1, 0} {
		if boots[i].Index != want {
			t.Errorf("boots[%d].Index = %d, want %d", i, boots[i].Index, want)
		}
	}

	// boot -1 is the one that froze on 9 Oct 2026: 18:49:28 → 12:11:56 the next day (CEST)
	b := boots[1]
	if b.ID != "2449e69564ba4aecbc8f807e0971b698" {
		t.Errorf("ID = %q", b.ID)
	}
	wantFirst := time.Date(2026, 10, 8, 16, 49, 28, 0, time.UTC) // 18:49:28 CEST
	if got := b.First.UTC().Truncate(time.Second); !got.Equal(wantFirst) {
		t.Errorf("First = %v, want %v", got, wantFirst)
	}
	if got := b.Duration().Round(time.Minute); got != 17*time.Hour+22*time.Minute {
		t.Errorf("Duration = %v, want 17h22m", got)
	}
}

func TestParseBootsInvalid(t *testing.T) {
	if _, err := parseBoots([]byte("not json")); err == nil {
		t.Error("expected an error for invalid JSON")
	}
}

func TestParseBootTable(t *testing.T) {
	// systemd 249 (no header) and 257 (header) print the same rows
	out := `IDX BOOT ID                          FIRST ENTRY                 LAST ENTRY
 -1 2449e69564ba4aecbc8f807e0971b698 Thu 2026-10-08 18:49:28 CEST Fri 2026-10-09 12:11:56 CEST
  0 7119bcd1612643c19756d30cebe2d2ff Fri 2026-10-09 12:12:59 CEST—Fri 2026-10-09 15:31:52 CEST
`
	boots := parseBootTable([]byte(out))
	if len(boots) != 2 || boots[0].Index != -1 || boots[0].ID != "2449e69564ba4aecbc8f807e0971b698" || boots[1].Index != 0 {
		t.Errorf("boots = %+v", boots)
	}
}
