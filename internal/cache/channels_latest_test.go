package cache

import "testing"

func TestAdvanceChannelLatestTS_NeverRegresses(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer db.Close()
	newRSChannel(t, db, "D1", "T1")

	for _, ts := range []string{"1700000002.000000", "1700000001.000000", ""} {
		if err := db.AdvanceChannelLatestTS("D1", ts); err != nil {
			t.Fatalf("AdvanceChannelLatestTS(%q): %v", ts, err)
		}
	}
	s, _ := db.GetChannelReadState("D1")
	if s.LatestTS != "1700000002.000000" {
		t.Errorf("LatestTS = %q, want 1700000002.000000", s.LatestTS)
	}
	ws, _ := db.GetWorkspaceReadState("T1")
	if ws["D1"].LatestTS != "1700000002.000000" {
		t.Errorf("workspace LatestTS = %q, want 1700000002.000000", ws["D1"].LatestTS)
	}
}

func TestReadStateWriters_AdvanceLatestTS(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer db.Close()
	newRSChannel(t, db, "D1", "T1")
	newRSChannel(t, db, "D2", "T1")

	if err := db.ReplaceWorkspaceReadState("T1", []ChannelReadStateUpdate{
		{ChannelID: "D1", LastReadTS: "1.0", LatestTS: "1700000005.000000"},
		{ChannelID: "D2", LatestTS: "1700000003.000000"},
	}); err != nil {
		t.Fatalf("ReplaceWorkspaceReadState: %v", err)
	}
	// Reconnect batch: D1 older value must not regress, D2 advances,
	// empty LatestTS leaves the column alone.
	if err := db.BatchUpdateChannelReadState([]ChannelReadStateUpdate{
		{ChannelID: "D1", LatestTS: "1700000004.000000"},
		{ChannelID: "D2", LatestTS: "1700000009.000000"},
	}); err != nil {
		t.Fatalf("BatchUpdateChannelReadState: %v", err)
	}
	ws, _ := db.GetWorkspaceReadState("T1")
	if got := ws["D1"].LatestTS; got != "1700000005.000000" {
		t.Errorf("D1 LatestTS = %q, want 1700000005.000000", got)
	}
	if got := ws["D2"].LatestTS; got != "1700000009.000000" {
		t.Errorf("D2 LatestTS = %q, want 1700000009.000000", got)
	}
}
