package config

import "testing"

func TestUIFeaturesConfigDefaults(t *testing.T) {
	f := UIFeaturesConfig{}.Resolved()
	for _, key := range []string{"graph", "kanban", "canvas", "whiteboard", "timeline", "calendar", "bases", "data_sources"} {
		if !f[key] {
			t.Fatalf("expected %s enabled by default", key)
		}
	}
}

func TestUIFeaturesConfigExplicitFalse(t *testing.T) {
	falseVal := false
	f := UIFeaturesConfig{
		Kanban:   &falseVal,
		Graph:    &falseVal,
		Calendar: &falseVal,
	}.Resolved()
	if f["kanban"] || f["graph"] || f["calendar"] {
		t.Fatal("expected kanban, graph, and calendar disabled")
	}
	if !f["canvas"] || !f["bases"] {
		t.Fatal("expected unset features to remain enabled")
	}
}

func TestUIFeaturesConfigIsEnabledUnknown(t *testing.T) {
	if !(UIFeaturesConfig{}).IsEnabled("unknown_feature") {
		t.Fatal("unknown feature names should default to enabled")
	}
}
