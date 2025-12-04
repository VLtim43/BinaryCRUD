package test

import (
	"BinaryCRUD/backend/search"
	"strings"
	"testing"
)

func TestBMBasicSearch(t *testing.T) {
	bm := search.NewBoyerMooreString("abc")
	matches := bm.SearchString("abcabcabc")

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

func TestBMNoMatch(t *testing.T) {
	bm := search.NewBoyerMooreString("xyz")
	matches := bm.SearchString("abcdefgh")

	if len(matches) != 0 {
		t.Errorf("Expected 0 matches, got %d", len(matches))
	}
}

func TestBMSingleChar(t *testing.T) {
	bm := search.NewBoyerMooreString("a")
	matches := bm.SearchString("abracadabra")

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

func TestBMOverlappingPatterns(t *testing.T) {
	bm := search.NewBoyerMooreString("aa")
	matches := bm.SearchString("aaaa")

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

func TestBMEmptyPattern(t *testing.T) {
	bm := search.NewBoyerMooreString("")
	matches := bm.SearchString("abc")

	if len(matches) != 0 {
		t.Errorf("Expected 0 matches for empty pattern, got %d", len(matches))
	}
}

func TestBMEmptyText(t *testing.T) {
	bm := search.NewBoyerMooreString("abc")
	matches := bm.SearchString("")

	if len(matches) != 0 {
		t.Errorf("Expected 0 matches for empty text, got %d", len(matches))
	}
}

func TestBMPatternLongerThanText(t *testing.T) {
	bm := search.NewBoyerMooreString("abcdefgh")
	matches := bm.SearchString("abc")

	if len(matches) != 0 {
		t.Errorf("Expected 0 matches when pattern > text, got %d", len(matches))
	}
}

func TestBMContains(t *testing.T) {
	bm := search.NewBoyerMooreString("needle")

	if !bm.ContainsString("haystackneedlehaystack") {
		t.Error("Expected to find 'needle' in text")
	}

	if bm.ContainsString("haystackhaystackhaystack") {
		t.Error("Should not find 'needle' in text without it")
	}
}

func TestBMConvenienceFunctions(t *testing.T) {
	matches := search.BMSearchString("abcabc", "abc")
	if len(matches) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(matches))
	}

	if !search.BMContainsString("hello world", "world") {
		t.Error("Expected to find 'world'")
	}

	if search.BMContainsString("hello world", "xyz") {
		t.Error("Should not find 'xyz'")
	}
}

func TestBMBytes(t *testing.T) {
	pattern := []byte{0x00, 0x01, 0x02}
	text := []byte{0xFF, 0x00, 0x01, 0x02, 0xFF, 0x00, 0x01, 0x02}

	matches := search.BMSearch(text, pattern)
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

func TestBMCaseInsensitive(t *testing.T) {
	pattern := "cola"
	text := "Cola"

	lowerPattern := strings.ToLower(pattern)
	lowerText := strings.ToLower(text)

	bm := search.NewBoyerMooreString(lowerPattern)
	result := bm.ContainsString(lowerText)

	if !result {
		t.Error("Expected to find 'cola' in 'cola'")
	}
}

func TestBMBadCharacterShift(t *testing.T) {
	// Test that bad character heuristic works correctly
	// Pattern "abc" in text "aabacabc" should find match at position 5
	bm := search.NewBoyerMooreString("abc")
	matches := bm.SearchString("aabacabc")

	if len(matches) != 1 || matches[0] != 5 {
		t.Errorf("Expected match at position 5, got %v", matches)
	}
}
