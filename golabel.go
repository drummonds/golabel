package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/drummonds/golabel/version"

	"github.com/mect/go-escpos"
	"golang.org/x/text/width"
)

//go:embed templates
var templateFS embed.FS

var p *escpos.Printer
var tmpl *template.Template

const maxLineLength = 32

// Command line flags
var (
	port = flag.Int("port", 80, "Port to listen on")
)

// max function for smart wrapping
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func spaceNearEnd(line []rune) (bool, int) {
	// Check if there's a space in the last 10 runes of the current line
	lastSpaceIndex := -1
	start := max(0, len(line)-10)

	// Look for the last space in the last 10 runes
	for j := len(line) - 1; j >= start; j-- {
		if unicode.IsSpace(line[j]) {
			lastSpaceIndex = j
			break
		}
	}
	result := lastSpaceIndex >= 0
	return result, lastSpaceIndex
}

// wrapSingleLine wraps a single line of text to fit within maxWidth
// maxWidth is the maximum width of the line in normal characters on
// an 80mm printer.  This should in time be transformed into a function
// in the driver that calculates it exactly
func wrapSingleLine(line string, maxPrintedWidth int) []string {
	if maxPrintedWidth <= 0 {
		return []string{line}
	}

	// Convert to runes and normalize for full-width characters
	runes := []rune(line)

	var lines []string
	var currentLine []rune
	currentWidth := 0

	// State machine for text wrapping
	type State int
	const (
		StateNormal State = iota
		StateSpaceSearch
		StateSkipSpace
	)

	currentState := StateSkipSpace // Start with skip space state
	var (
		currentRune    rune
		charWidth      int
		lastSpaceIndex int
		hasSpace       bool
	)

	StateWrapAtSpace := func(lastSpaceIndex int) {
		currentLine = append(currentLine, currentRune)
		// Split at the space and consume it
		beforeSpace := currentLine[:lastSpaceIndex]
		afterSpace := currentLine[lastSpaceIndex+1:]

		lines = append(lines, strings.TrimSpace(string(beforeSpace)))
		currentLine = afterSpace
		currentWidth = runesWidth(currentLine)
		if currentWidth == 0 {
			currentState = StateSkipSpace // Start new line with skip space state
			return
		}
		currentState = StateNormal
		return // Start new line with skip space state
	}
	StateWrapAtCharacter := func() {
		// No space found, wrap at current position and add hyphen
		lines = append(lines, strings.TrimSpace(string(currentLine[:len(currentLine)-1]))+"-")
		currentLine = []rune{currentLine[len(currentLine)-1], currentRune}
		currentWidth = runesWidth(currentLine)
		currentState = StateNormal
	}

	for i := 0; i < len(runes); i++ {
		currentRune = runes[i]
		charWidth = runeWidth(currentRune)

		switch currentState {
		case StateSkipSpace:
			// Skip leading spaces at the beginning of a line
			if unicode.IsSpace(currentRune) {
				// Continue skipping spaces
				continue
			} else {
				// Found non-space character, transition to normal state
				currentState = StateNormal
				// Don't consume the rune yet, re-process it in normal state
				i-- // Rewind to reprocess this rune
			}

		case StateNormal:
			// Check if adding this character would exceed the line width
			if currentWidth+charWidth > maxPrintedWidth && currentWidth > 0 {
				switch {
				// case i == len(runes)-1: // end of line
				// 	StateWrapAtCharacter()
				case unicode.IsSpace(currentRune): // next char is a space
					StateWrapAtSpace(len(currentLine))
				default:
					hasSpace, lastSpaceIndex = spaceNearEnd(currentLine)
					if hasSpace {
						StateWrapAtSpace(lastSpaceIndex)
					} else {
						StateWrapAtCharacter()
					}
				}
			} else {
				// Add the character to the current line
				currentLine = append(currentLine, currentRune)
				currentWidth += charWidth
			}
		}
	}

	// Add the last line if it has content
	if len(currentLine) > 0 {
		lines = append(lines, strings.TrimSpace(string(currentLine)))
	}

	// Handle empty input
	if len(lines) == 0 {
		return []string{""}
	}

	return lines
}

// wrapTextUnicode is a Unicode-aware text wrapping function
// maxWidth is the maximum width of the line in normal characters on
// an 80mm printer.  This should in time be transformed into a function
// in the driver that
func wrapTextUnicode(text string, maxWidth int) []string {
	if maxWidth <= 0 { // no wrapping
		return []string{text}
	}

	var allLines []string

	// Split the text by newlines and process each line separately
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			allLines = append(allLines, "") // Preserve empty lines
			continue
		}

		// Wrap this single line
		wrappedLines := wrapSingleLine(line, maxWidth)
		allLines = append(allLines, wrappedLines...)
	}

	// Handle empty input
	if len(allLines) == 0 {
		return []string{""}
	}

	return allLines
}

// runeWidth returns the display width of a rune
func runeWidth(r rune) int {
	switch {
	case r == '\t':
		return 4 // Tab width
	case r == '\n':
		return 0 // Newlines don't take width
	case unicode.IsControl(r):
		return 0 // Control characters don't take width
	case width.LookupRune(r).Kind() == width.EastAsianWide:
		return 2 // Full-width characters
	default:
		return 1 // Regular characters
	}
}

// runeWidth returns the display width of a rune
func runesWidth(runes []rune) (result int) {
	for _, thisRune := range runes {
		result += runeWidth(thisRune)
	}
	return result
}

// printMessageLines prints a message line by line, wrapping at maxLineLength characters
func printMessageLines(message string) {
	message = strings.TrimSpace(message)
	if message == "" {
		return
	}

	// Split by line breaks and process each line
	lines := strings.Split(message, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			p.PrintLn("") // Print blank line for empty lines
			continue
		}

		// Use the Unicode-aware wrapping function
		wrappedLines := wrapTextUnicode(line, maxLineLength)
		for _, wrappedLine := range wrappedLines {
			p.PrintLn(wrappedLine)
		}
	}
}

func label(message string, num int) error {
	if p == nil {
		return fmt.Errorf("printer not initialized")
	}

	p.Init()       // start
	p.Smooth(true) // use smooth printing
	p.Size(3, 3)   // set font size
	p.Underline(true)
	p.Align(escpos.AlignCenter)
	p.PrintLn("Task")
	p.Underline(false)

	p.Size(2, 2)
	p.Font(escpos.FontB) // change font
	p.Align(escpos.AlignLeft)

	// Sanitize message to prevent injection
	message = strings.TrimSpace(message)
	if message == "" {
		return fmt.Errorf("message cannot be empty")
	}

	printMessageLines(message)

	p.Feed(2)

	p.Align(escpos.AlignCenter)
	p.Barcode(fmt.Sprintf("%d", num), escpos.BarcodeTypeCODE39) // print barcode
	p.Align(escpos.AlignLeft)
	p.Size(1, 1) // set font size
	p.PrintLn("Printed at: " + time.Now().Format("2006-01-02T15:04:05Z"))

	p.Cut() // cut
	p.End() // stop
	time.Sleep(time.Second * 1)
	return nil
}

type PageData struct {
	Status    string
	Success   bool
	Version   string
	BuildDate string
}

func handlePrint(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		message := r.FormValue("message")
		barcodeStr := r.FormValue("barcode")

		if message == "" {
			renderPage(w, "Error: Message cannot be empty", false)
			return
		}

		barcode, err := strconv.Atoi(barcodeStr)
		if err != nil {
			barcode = 5 // default value
		}

		// Call the label function
		err = label(message, barcode)
		if err != nil {
			renderPage(w, err.Error(), false)
			return
		}

		renderPage(w, "Label printed successfully!", true)
	} else {
		renderPage(w, "", false)
	}
}

func renderPage(w http.ResponseWriter, status string, success bool) {
	if tmpl == nil {
		http.Error(w, "Template not initialized", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Status:    status,
		Success:   success,
		Version:   version.GetVersion(),
		BuildDate: version.GetBuildDate(),
	}

	err := tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	// Parse command line flags
	flag.Parse()

	var err error

	// Initialize template from embedded filesystem
	tmpl, err = template.ParseFS(templateFS, "templates/printer.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}

	p, err = escpos.NewUSBPrinterByPath("") // empty string will do a self discovery
	if err != nil {
		fmt.Println("Error initializing printer:", err)
		return
	}

	fmt.Printf("Starting GoLabel web server on http://localhost:%d\n", *port)
	fmt.Printf("Version: %s\n", version.GetVersionInfo())
	fmt.Println("Printer initialized successfully")

	http.HandleFunc("/", handlePrint)
	http.HandleFunc("/print", handlePrint)

	err = http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
