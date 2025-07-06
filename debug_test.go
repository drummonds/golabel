package main

import (
	"fmt"
	"testing"
)

func TestDebugWrapTextUnicode(t *testing.T) {
	text := "This is a very long sentence that should wrap at spaces"
	maxWidth := 20

	fmt.Printf("Input: %q\n", text)
	fmt.Printf("Max width: %d\n", maxWidth)

	// Calculate character widths
	var currentWidth int
	for i, r := range text {
		charWidth := runeWidth(r)
		currentWidth += charWidth
		fmt.Printf("Char %d: %q (width: %d, total: %d)\n", i, r, charWidth, currentWidth)
	}
	fmt.Printf("Total display width: %d\n", currentWidth)

	result := wrapTextUnicode(text, maxWidth)
	fmt.Printf("Result: %v\n", result)
}
