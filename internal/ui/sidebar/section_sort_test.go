package sidebar

import (
	"testing"

	"github.com/gammons/slk/internal/core"
)

func visibleIDs(m *Model) []string {
	var ids []string
	for _, it := range m.VisibleItems() {
		ids = append(ids, it.ID)
	}
	return ids
}

func assertIDs(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

// Slack's "default" channel_sort renders non-DM sections A–Z by display
// name, case-insensitively. users.conversations returns channels in an
// arbitrary order, so input order must not leak into the sidebar.
func TestSlackMode_NonDMSectionsSortAlphabetically(t *testing.T) {
	items := []ChannelItem{
		{ID: "C3", Name: "zeta", Type: "channel", Section: "PROJ"},
		{ID: "S2", Name: "general", Type: "channel", Section: "STARS"},
		{ID: "C1", Name: "Alpha", Type: "private", Section: "PROJ"},
		{ID: "S1", Name: "bob", Type: "dm", Section: "STARS"},
		{ID: "C2", Name: "beta", Type: "channel", Section: "PROJ"},
		{ID: "C5", Name: "random", Type: "channel"}, // type-default: Channels
		{ID: "C4", Name: "announce", Type: "channel"},
	}
	provider := &fakeProvider{
		ready: true,
		sections: []SectionMeta{
			{ID: "STARS", Name: "", Type: "stars"},
			{ID: "CH", Name: "Channels", Type: "channels"},
			{ID: "PROJ", Name: "Ongoing Projects", Type: "standard"},
		},
	}
	m := New(items)
	m.SetSectionsProvider(provider)
	assertIDs(t, visibleIDs(&m), []string{"S1", "S2", "C4", "C5", "C1", "C2", "C3"})
}

// Config-glob mode keeps its documented behavior: un-annotated channels
// stay in input order.
func TestConfigMode_UnannotatedKeepsInputOrder(t *testing.T) {
	items := []ChannelItem{
		{ID: "C2", Name: "zeta", Type: "channel", Section: "Eng", SectionOrder: 1},
		{ID: "C1", Name: "alpha", Type: "channel", Section: "Eng", SectionOrder: 1},
	}
	m := New(items)
	assertIDs(t, visibleIDs(&m), []string{"C2", "C1"})
}

func dmSortFixture() (Model, map[string]core.ReadState) {
	items := []ChannelItem{
		{ID: "D1", Name: "alice", Type: "dm"},
		{ID: "D2", Name: "bob", Type: "dm"},
		{ID: "D3", Name: "carol", Type: "dm"},
		{ID: "G1", Name: "dave, erin", Type: "group_dm"},
	}
	state := map[string]core.ReadState{
		"D1": {LastReadTS: "1700000001.000000", LatestTS: "1700000001.000000"},
		"D2": {LastReadTS: "1700000002.000000", LatestTS: "1700000300.000000"},
		// No LatestTS yet: falls back to LastReadTS.
		"D3": {LastReadTS: "1700000150.000000"},
		"G1": {LastReadTS: "1700000002.000000", LatestTS: "1700000200.000000"},
	}
	m := New(items)
	m.SetReadStateReader(func() map[string]core.ReadState { return state })
	m.SetSectionsProvider(&fakeProvider{
		ready:    true,
		sections: []SectionMeta{{ID: "DMS", Name: "Direct Messages", Type: "direct_messages"}},
	})
	return m, state
}

// Slack's "default" channel_sort renders Direct Messages most recent
// first.
func TestSlackMode_DirectMessagesSortByRecency(t *testing.T) {
	m, _ := dmSortFixture()
	assertIDs(t, visibleIDs(&m), []string{"D2", "G1", "D3", "D1"})
}

// A new message in a DM re-sorts the section live.
func TestSlackMode_NoteActivityResortsDMs(t *testing.T) {
	m, state := dmSortFixture()
	state["D1"] = core.ReadState{LastReadTS: "1700000001.000000", LatestTS: "1700000400.000000"}
	m.NoteActivity("D1")
	assertIDs(t, visibleIDs(&m), []string{"D1", "D2", "G1", "D3"})
}
