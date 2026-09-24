package sidebar

import "testing"

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
