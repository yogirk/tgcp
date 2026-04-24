package core

import (
	"strings"
	"testing"
)

// TestFilterCommands_PrefixBeatsFuzzy is the regression test for the
// "gce" → GKE-ranked-first bug. The ranker must prefer literal prefix
// matches on Name over a fuzzy hit spread across Name + Description.
func TestFilterCommands_PrefixBeatsFuzzy(t *testing.T) {
	m := NewNavigation()
	m.FilterCommands("gce")

	if len(m.Suggestions) == 0 {
		t.Fatal("expected at least one suggestion for 'gce'")
	}
	first := m.Suggestions[0].Name
	if !strings.HasPrefix(strings.ToLower(first), "gce") {
		t.Fatalf("expected first result to start with 'gce', got %q", first)
	}
}

func TestFilterCommands_SubstringBeatsFuzzy(t *testing.T) {
	// "cluster" is not a prefix of any command; it is a literal substring
	// of "GKE: List Clusters" and "Dataproc: List Clusters". Those should
	// outrank any fuzzy match that only shares individual characters.
	m := NewNavigation()
	m.FilterCommands("cluster")

	if len(m.Suggestions) < 2 {
		t.Fatalf("expected ≥2 suggestions for 'cluster', got %d", len(m.Suggestions))
	}
	for i, s := range m.Suggestions[:2] {
		if !strings.Contains(strings.ToLower(s.Name), "cluster") {
			t.Errorf("suggestion[%d] %q missing literal 'cluster' substring", i, s.Name)
		}
	}
}

func TestFilterCommands_EmptyQueryClearsSuggestions(t *testing.T) {
	m := NewNavigation()
	m.FilterCommands("gce")
	if len(m.Suggestions) == 0 {
		t.Fatal("setup: expected matches for 'gce'")
	}
	m.FilterCommands("")
	if len(m.Suggestions) != 0 {
		t.Errorf("empty query should clear suggestions, got %d", len(m.Suggestions))
	}
}

func TestFilterCommands_NoMatch(t *testing.T) {
	m := NewNavigation()
	m.FilterCommands("zzzz-definitely-no-match-zzzz")
	if len(m.Suggestions) != 0 {
		t.Errorf("expected 0 suggestions for impossible query, got %d", len(m.Suggestions))
	}
}

func TestFilterCommands_MatchedIndexesCoverQuery(t *testing.T) {
	// For prefix and substring tiers, MatchedIndexes must cover the entire
	// query span so the palette highlights the whole match, not just the
	// first char.
	m := NewNavigation()
	m.FilterCommands("gcs")
	if len(m.Suggestions) == 0 {
		t.Fatal("expected matches for 'gcs'")
	}
	first := m.Suggestions[0]
	if len(first.MatchedIndexes) != 3 {
		t.Errorf("expected 3 highlighted chars for 'gcs', got %d", len(first.MatchedIndexes))
	}
}
