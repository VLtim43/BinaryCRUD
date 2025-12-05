package search

// KMP implements the Knuth-Morris-Pratt string matching algorithm.
// Time complexity: O(n + m) where n is text length and m is pattern length.
type KMP struct {
	pattern []byte
	lps     []int // Longest Proper Prefix which is also Suffix
}

// NewKMP creates a new KMP instance with the given pattern.
// Precomputes the LPS array for efficient searching.
func NewKMP(pattern []byte) *KMP {
	kmp := &KMP{
		pattern: pattern,
		lps:     computeLPS(pattern),
	}
	return kmp
}

// NewKMPString creates a new KMP instance from a string pattern.
func NewKMPString(pattern string) *KMP {
	return NewKMP([]byte(pattern))
}

// computeLPS builds the Longest Proper Prefix Suffix array.
// LPS[i] = length of the longest proper prefix of pattern[0..i]
// which is also a suffix of pattern[0..i].
func computeLPS(pattern []byte) []int {
	m := len(pattern)
	if m == 0 {
		return []int{}
	}

	lps := make([]int, m)
	lps[0] = 0 // LPS of single char is always 0

	length := 0 // Length of previous longest prefix suffix
	i := 1

	for i < m {
		if pattern[i] == pattern[length] {
			length++
			lps[i] = length
			i++
		} else {
			if length != 0 {
				// Use previously computed LPS value
				length = lps[length-1]
			} else {
				lps[i] = 0
				i++
			}
		}
	}

	return lps
}

// Search finds all occurrences of the pattern in the text.
// Returns a slice of starting indices where the pattern was found.
func (k *KMP) Search(text []byte) []int {
	var matches []int

	n := len(text)
	m := len(k.pattern)

	if m == 0 || n == 0 || m > n {
		return matches
	}

	i := 0 // Index for text
	j := 0 // Index for pattern

	for i < n {
		if k.pattern[j] == text[i] {
			i++
			j++
		}

		if j == m {
			// Found a match at index i-j
			matches = append(matches, i-j)
			j = k.lps[j-1]
		} else if i < n && k.pattern[j] != text[i] {
			if j != 0 {
				j = k.lps[j-1]
			} else {
				i++
			}
		}
	}

	return matches
}

// SearchString searches for the pattern in a string text.
func (k *KMP) SearchString(text string) []int {
	return k.Search([]byte(text))
}

// Contains returns true if the pattern exists in the text.
func (k *KMP) Contains(text []byte) bool {
	n := len(text)
	m := len(k.pattern)

	if m == 0 || n == 0 || m > n {
		return false
	}

	i := 0
	j := 0

	for i < n {
		if k.pattern[j] == text[i] {
			i++
			j++
		}

		if j == m {
			return true
		} else if i < n && k.pattern[j] != text[i] {
			if j != 0 {
				j = k.lps[j-1]
			} else {
				i++
			}
		}
	}

	return false
}

// ContainsString returns true if the pattern exists in the string text.
func (k *KMP) ContainsString(text string) bool {
	return k.Contains([]byte(text))
}

// Search finds all occurrences of pattern in text.
func Search(text, pattern []byte) []int {
	kmp := NewKMP(pattern)
	return kmp.Search(text)
}

// SearchString finds all occurrences of pattern in text (string version).
func SearchString(text, pattern string) []int {
	kmp := NewKMPString(pattern)
	return kmp.SearchString(text)
}

// Contains returns true if pattern exists in text.
func Contains(text, pattern []byte) bool {
	kmp := NewKMP(pattern)
	return kmp.Contains(text)
}

// ContainsString returns true if pattern exists in text (string version).
func ContainsString(text, pattern string) bool {
	kmp := NewKMPString(pattern)
	return kmp.ContainsString(text)
}
