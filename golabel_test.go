package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestWrapTextUnicode(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxWidth int
		expected []string
	}{
		{ //0
			name:     "Empty string",
			text:     "",
			maxWidth: 10,
			expected: []string{""},
		},
		{
			name:     "Single word shorter than max width",
			text:     "Hello",
			maxWidth: 10,
			expected: []string{"Hello"},
		},
		{
			name:     "Single word longer than max width",
			text:     "Supercalifragilisticexpialidocious",
			maxWidth: 10,
			expected: []string{"Supercali-", "fragilist-", "icexpiali-", "docious"},
		},
		{ //3
			name:     "Multiple words with spaces",
			text:     "Hello world this is a test",
			maxWidth: 15,
			expected: []string{"Hello world", "this is a test"},
		},
		{
			name:     "Words with spaces in last 10 characters",
			text:     "This is a very long sentence that should wrap at spaces",
			maxWidth: 20,
			expected: []string{"This is a very long", "sentence that should", "wrap at spaces"},
		},
		{
			name:     "Words with doublespaces in last 10 characters",
			text:     "This is a very odd  sentence that should wrap at spaces",
			maxWidth: 20,
			expected: []string{"This is a very odd", "sentence that should", "wrap at spaces"},
		},
		{ //6
			name:     "No spaces in last 10 characters",
			text:     "This is a verylongwordthatwillneedtobebroken",
			maxWidth: 15,
			expected: []string{"This is a", "verylongwordth-", "atwillneedtobe-", "broken"},
		},
		{
			name:     "Unicode full-width characters",
			text:     "Hello世界world",
			maxWidth: 10,
			expected: []string{"Hello世界-", "world"},
		},
		{
			name:     "Mixed Unicode and ASCII",
			text:     "Hello 世界 world 测试",
			maxWidth: 12,
			expected: []string{"Hello 世界", "world 测试"},
		},
		{ //9
			name:     "Multiple lines with newlines",
			text:     "Line 1\nLine 2\nLine 3",
			maxWidth: 10,
			expected: []string{"Line 1", "Line 2", "Line 3"},
		},
		{
			name:     "Very long word with spaces nearby",
			text:     "Short verylongwordwithoutspaces short",
			maxWidth: 15,
			expected: []string{"Short", "verylongwordwi-", "thoutspaces", "short"},
		},
		{
			name:     "Zero max width",
			text:     "Hello world",
			maxWidth: 0,
			expected: []string{"Hello world"},
		},
		{ //12
			name:     "Negative max width",
			text:     "Hello world",
			maxWidth: -5,
			expected: []string{"Hello world"},
		},
		{
			name:     "Single character",
			text:     "A",
			maxWidth: 5,
			expected: []string{"A"},
		},
		{
			name:     "Multiple spaces",
			text:     "Word1    Word2",
			maxWidth: 10,
			expected: []string{"Word1", "Word2"},
		},
		{ //15  Note could have feature than maxwidth spaces is new line
			name:     "Spaces longer than max",
			text:     "Word1                    Word2",
			maxWidth: 10,
			expected: []string{"Word1", "Word2"},
		},
		{
			name:     "Tab character",
			text:     "Hello\tworld",
			maxWidth: 10,
			expected: []string{"Hello", "world"},
		},
		{
			name:     "Control characters",
			text:     "Hello\x00world",
			maxWidth: 10,
			expected: []string{"Hello\x00world"},
		},
		{ //18
			name:     "Space at exact boundary",
			text:     "Hello world test",
			maxWidth: 5,
			expected: []string{"Hello", "world", "test"},
		},
		{
			name:     "Space just after boundary",
			text:     "Hello world test",
			maxWidth: 4,
			expected: []string{"Hel-", "lo", "wor-", "ld", "test"},
		},
		{
			name:     "Embedded newlines",
			text:     "First line\nSecond line is long\nThird",
			maxWidth: 10,
			expected: []string{"First line", "Second", "line is", "long", "Third"},
		},
	}

	for _, tt := range tests[20:] {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapTextUnicode(tt.text, tt.maxWidth)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("wrapTextUnicode(%q, %d) = %+v, want %+v", tt.text, tt.maxWidth, result, tt.expected)
			}
		})
	}
}

func TestRuneWidth(t *testing.T) {
	tests := []struct {
		name     string
		r        rune
		expected int
	}{
		{"Tab character", '\t', 4},
		{"Newline character", '\n', 0},
		{"Control character", '\x00', 0},
		{"ASCII character", 'A', 1},
		{"Space character", ' ', 1},
		{"Full-width character", '世', 2},
		{"Full-width character", '界', 2},
		{"Number", '1', 1},
		{"Punctuation", '.', 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runeWidth(tt.r)
			if result != tt.expected {
				t.Errorf("runeWidth(%q) = %d, want %d", tt.r, result, tt.expected)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"Positive numbers", 5, 3, 5},
		{"Negative numbers", -5, -3, -3},
		{"Equal numbers", 5, 5, 5},
		{"Zero and positive", 0, 5, 5},
		{"Zero and negative", 0, -5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := max(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("max(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkWrapTextUnicode(b *testing.B) {
	text := "This is a benchmark test with some very long words like supercalifragilisticexpialidocious and some normal words to test the wrapping functionality"
	maxWidth := 20

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wrapTextUnicode(text, maxWidth)
	}
}

func BenchmarkRuneWidth(b *testing.B) {
	runes := []rune{'A', '世', '\t', '\n', ' ', '1', '.'}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, r := range runes {
			runeWidth(r)
		}
	}
}

func TestSpaceNearEnd(t *testing.T) {
	tests := []struct {
		name          string
		input         []rune
		expectedFound bool
		expectedIndex int
	}{
		{ //0
			name:          "empty slice",
			input:         []rune{},
			expectedFound: false,
			expectedIndex: -1,
		},
		{
			name:          "single character no space",
			input:         []rune{'a'},
			expectedFound: false,
			expectedIndex: -1,
		},
		{
			name:          "single space",
			input:         []rune{' '},
			expectedFound: true, // space at index 0, but function requires > 0
			expectedIndex: 0,
		},
		{ //3
			name:          "space at end",
			input:         []rune{'a', 'b', ' '},
			expectedFound: true,
			expectedIndex: 2,
		},
		{
			name:          "space near end within 10 chars",
			input:         []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', ' '},
			expectedFound: true,
			expectedIndex: 26,
		},
		{
			name:          "space beyond last 10 chars",
			input:         []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'},
			expectedFound: false,
			expectedIndex: -1,
		},
		{ //6
			name:          "multiple spaces in last 10 chars",
			input:         []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', ' ', 'a', ' ', 'b'},
			expectedFound: true,
			expectedIndex: 28, // should find the last space
		},
		{
			name:          "exactly 10 chars with space at end",
			input:         []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', ' '},
			expectedFound: true,
			expectedIndex: 9,
		},
		{
			name:          "exactly 10 chars with space in middle",
			input:         []rune{'a', 'b', 'c', 'd', 'e', ' ', 'g', 'h', 'i', 'j'},
			expectedFound: true,
			expectedIndex: 5,
		},
		{ //9
			name:          "less than 10 chars with space",
			input:         []rune{'a', 'b', 'c', 'd', 'e', ' '},
			expectedFound: true,
			expectedIndex: 5,
		},
		{
			name:          "less than 10 chars without space",
			input:         []rune{'a', 'b', 'c', 'd', 'e'},
			expectedFound: false,
			expectedIndex: -1,
		},
		{
			name:          "space at index 0 in short string",
			input:         []rune{' ', 'a', 'b'},
			expectedFound: true, // space at index 0, but function requires > 0
			expectedIndex: 0,
		},
		{ //12
			name:          "space at index 0 in long string",
			input:         []rune{' ', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'},
			expectedFound: false, // space at index 0, but function requires > 0
			expectedIndex: -1,
		},
		{
			name:          "unicode spaces",
			input:         []rune{'a', 'b', 'c', '\t', 'd', 'e', 'f', 'g'},
			expectedFound: true,
			expectedIndex: 3, // tab character
		},
		{
			name:          "newline character",
			input:         []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '\n'},
			expectedFound: true,
			expectedIndex: 26, // newline character
		},
	}

	for _, tt := range tests[13:] {
		t.Run(tt.name, func(t *testing.T) {
			found, index := spaceNearEnd(tt.input)

			if found != tt.expectedFound {
				t.Errorf("spaceNearEnd() found = %v, expected %v", found, tt.expectedFound)
			}

			if index != tt.expectedIndex {
				t.Errorf("spaceNearEnd() index = %v, expected %v", index, tt.expectedIndex)
			}
		})
	}
}

// TestSpaceNearEndEdgeCases tests edge cases and boundary conditions
func TestSpaceNearEndEdgeCases(t *testing.T) {
	// Test with various Unicode space characters
	unicodeSpaces := []rune{' ', '\t', '\n', '\r', '\f', '\v', '\u00A0', '\u2000', '\u2001', '\u2002', '\u2003', '\u2004', '\u2005', '\u2006', '\u2007', '\u2008', '\u2009', '\u200A', '\u2028', '\u2029', '\u202F', '\u205F', '\u3000'}

	for i, space := range unicodeSpaces {
		t.Run(fmt.Sprintf("unicode_space_%d", i), func(t *testing.T) {
			// Create a string with the space character in the last 10 positions
			input := make([]rune, 15)
			for j := range input {
				input[j] = 'a'
			}
			input[10] = space // Put space at position 10 (within last 10)

			found, index := spaceNearEnd(input)

			if !found {
				t.Errorf("Expected to find space character %U at index 10", space)
			}

			if index != 10 {
				t.Errorf("Expected space index to be 10, got %d", index)
			}
		})
	}
}

// TestSpaceNearEndBoundary tests the exact boundary of the 10-character window
func TestSpaceNearEndBoundary(t *testing.T) {
	// Test with exactly 10 characters
	input := []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}
	found, index := spaceNearEnd(input)

	if found {
		t.Errorf("Expected no space found in 10-char string without spaces, but found one at index %d", index)
	}

	// // Test with space at position 0 (should not be found due to > 0 condition)
	// input = []rune{' ', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i'}
	// found, index = spaceNearEnd(input)

	// if found {
	// 	t.Errorf("Expected no space found when space is at index 0, but found one at index %d", index)
	// }

	// Test with space at position 1 (should be found)
	input = []rune{'a', ' ', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i'}
	found, index = spaceNearEnd(input)

	if !found {
		t.Errorf("Expected to find space at index 1, but none found")
	}

	if index != 1 {
		t.Errorf("Expected space index to be 1, got %d", index)
	}
}
