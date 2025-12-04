package test

import (
	"BinaryCRUD/backend/search"
	"strings"
	"testing"
)

func TestKMPBasicSearch(t *testing.T) {
	kmp := search.NewKMPString("abc")
	matches := kmp.SearchString("abcabcabc")

	expected := []int{0, 3, 6}
	if len(matches) != len(expected) {
		t.Errorf("Expected %d matches, got %d", len(expected), len(matches))
		return
	}

	for i, v := range matches {
		if v != expected[i] {
			t.Errorf("Match %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestKMPNoMatch(t *testing.T) {
	kmp := search.NewKMPString("xyz")
	matches := kmp.SearchString("abcdefgh")

	if len(matches) != 0 {
		t.Errorf("Expected 0 matches, got %d", len(matches))
	}
}

func TestKMPSingleChar(t *testing.T) {
	kmp := search.NewKMPString("a")
	matches := kmp.SearchString("abracadabra")

	expected := []int{0, 3, 5, 7, 10}
	if len(matches) != len(expected) {
		t.Errorf("Expected %d matches, got %d", len(expected), len(matches))
		return
	}

	for i, v := range matches {
		if v != expected[i] {
			t.Errorf("Match %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestKMPOverlappingPatterns(t *testing.T) {
	kmp := search.NewKMPString("aa")
	matches := kmp.SearchString("aaaa")

	// "aa" appears at positions 0, 1, 2 (overlapping)
	expected := []int{0, 1, 2}
	if len(matches) != len(expected) {
		t.Errorf("Expected %d matches, got %d", len(expected), len(matches))
		return
	}

	for i, v := range matches {
		if v != expected[i] {
			t.Errorf("Match %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestKMPEmptyPattern(t *testing.T) {
	kmp := search.NewKMPString("")
	matches := kmp.SearchString("abc")

	if len(matches) != 0 {
		t.Errorf("Expected 0 matches for empty pattern, got %d", len(matches))
	}
}

func TestKMPEmptyText(t *testing.T) {
	kmp := search.NewKMPString("abc")
	matches := kmp.SearchString("")

	if len(matches) != 0 {
		t.Errorf("Expected 0 matches for empty text, got %d", len(matches))
	}
}

func TestKMPPatternLongerThanText(t *testing.T) {
	kmp := search.NewKMPString("abcdefgh")
	matches := kmp.SearchString("abc")

	if len(matches) != 0 {
		t.Errorf("Expected 0 matches when pattern > text, got %d", len(matches))
	}
}

func TestKMPContains(t *testing.T) {
	kmp := search.NewKMPString("needle")

	if !kmp.ContainsString("haystackneedlehaystack") {
		t.Error("Expected to find 'needle' in text")
	}

	if kmp.ContainsString("haystackhaystackhaystack") {
		t.Error("Should not find 'needle' in text without it")
	}
}

func TestKMPConvenienceFunctions(t *testing.T) {
	matches := search.SearchString("abcabc", "abc")
	if len(matches) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(matches))
	}

	if !search.ContainsString("hello world", "world") {
		t.Error("Expected to find 'world'")
	}

	if search.ContainsString("hello world", "xyz") {
		t.Error("Should not find 'xyz'")
	}
}

func TestKMPBytes(t *testing.T) {
	pattern := []byte{0x00, 0x01, 0x02}
	text := []byte{0xFF, 0x00, 0x01, 0x02, 0xFF, 0x00, 0x01, 0x02}

	matches := search.Search(text, pattern)
	expected := []int{1, 5}

	if len(matches) != len(expected) {
		t.Errorf("Expected %d matches, got %d", len(expected), len(matches))
		return
	}

	for i, v := range matches {
		if v != expected[i] {
			t.Errorf("Match %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestKMPLPSComputation(t *testing.T) {
	// Test with pattern "AAACAAAA"
	// LPS should be [0, 1, 2, 0, 1, 2, 3, 3]
	kmp := search.NewKMPString("AAACAAAA")
	matches := kmp.SearchString("AAACAAAAB")

	if len(matches) != 1 || matches[0] != 0 {
		t.Errorf("Expected match at 0, got %v", matches)
	}
}

func TestSearchCola(t *testing.T) {
	pattern := "cola"
	text := "Cola"
	
	lowerPattern := strings.ToLower(pattern)
	lowerText := strings.ToLower(text)
	
	kmp := search.NewKMPString(lowerPattern)
	result := kmp.ContainsString(lowerText)
	
	t.Logf("Pattern: %q, Text: %q, Result: %v", lowerPattern, lowerText, result)
	if !result {
		t.Error("Expected to find 'cola' in 'cola'")
	}
}
