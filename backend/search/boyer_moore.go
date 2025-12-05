package search

// BoyerMoore implements the Boyer-Moore string matching algorithm.
// Time complexity: O(n/m) best case, O(n*m) worst case
type BoyerMoore struct {
	pattern      []byte
	badCharTable [256]int // Bad character shift table
}

// NewBoyerMoore creates a new BoyerMoore instance with the given pattern.
// Precomputes the bad character table for efficient searching.
func NewBoyerMoore(pattern []byte) *BoyerMoore {
	bm := &BoyerMoore{
		pattern: pattern,
	}
	bm.computeBadCharTable()
	return bm
}

func NewBoyerMooreString(pattern string) *BoyerMoore {
	return NewBoyerMoore([]byte(pattern))
}

// computeBadCharTable builds the bad character shift table.
// For each character, stores the rightmost position in the pattern.
// Characters not in pattern get a shift of pattern length.
func (bm *BoyerMoore) computeBadCharTable() {
	m := len(bm.pattern)

	// Initialize all characters with pattern length (max shift)
	for i := 0; i < 256; i++ {
		bm.badCharTable[i] = m
	}

	// For characters in pattern, store distance from rightmost occurrence to end
	// We don't include the last character to avoid zero shifts
	for i := 0; i < m-1; i++ {
		bm.badCharTable[bm.pattern[i]] = m - 1 - i
	}
}

// Search finds all occurrences of the pattern in the text.
// Returns a slice of starting indices where the pattern was found.
func (bm *BoyerMoore) Search(text []byte) []int {
	var matches []int

	n := len(text)
	m := len(bm.pattern)

	if m == 0 || n == 0 || m > n {
		return matches
	}

	// Start comparing from the end of the pattern
	i := 0 // Position in text where pattern starts

	for i <= n-m {
		j := m - 1 // Start from end of pattern

		// Compare pattern from right to left
		for j >= 0 && bm.pattern[j] == text[i+j] {
			j--
		}

		if j < 0 {
			// Pattern found at position i
			matches = append(matches, i)
			// Shift by 1 to find overlapping matches
			i++
		} else {
			badCharShift := bm.badCharTable[text[i+j]]

			// Calculate shift: align the bad character or shift by 1
			shift := badCharShift - (m - 1 - j)
			if shift < 1 {
				shift = 1
			}
			i += shift
		}
	}

	return matches
}

// SearchString searches for the pattern in a string text.
func (bm *BoyerMoore) SearchString(text string) []int {
	return bm.Search([]byte(text))
}

func (bm *BoyerMoore) Contains(text []byte) bool {
	n := len(text)
	m := len(bm.pattern)

	if m == 0 || n == 0 || m > n {
		return false
	}

	i := 0

	for i <= n-m {
		j := m - 1

		for j >= 0 && bm.pattern[j] == text[i+j] {
			j--
		}

		if j < 0 {
			return true
		}

		badCharShift := bm.badCharTable[text[i+j]]
		shift := badCharShift - (m - 1 - j)
		if shift < 1 {
			shift = 1
		}
		i += shift
	}

	return false
}

// ContainsString returns true if the pattern exists in the string text.
func (bm *BoyerMoore) ContainsString(text string) bool {
	return bm.Contains([]byte(text))
}

// BMSearch finds all occurrences of pattern in text using Boyer-Moore.
func BMSearch(text, pattern []byte) []int {
	bm := NewBoyerMoore(pattern)
	return bm.Search(text)
}

// BMSearchString finds all occurrences of pattern in text (string version).
func BMSearchString(text, pattern string) []int {
	bm := NewBoyerMooreString(pattern)
	return bm.SearchString(text)
}

// BMContains returns true if pattern exists in text using Boyer-Moore.
func BMContains(text, pattern []byte) bool {
	bm := NewBoyerMoore(pattern)
	return bm.Contains(text)
}

// BMContainsString returns true if pattern exists in text (string version).
func BMContainsString(text, pattern string) bool {
	bm := NewBoyerMooreString(pattern)
	return bm.ContainsString(text)
}
