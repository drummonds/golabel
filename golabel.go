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

const maxLineLength = 27

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

// wrapSingleLine wraps a single line of text to fit within maxWidth
func wrapSingleLine(line string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{line}
	}

	var lines []string
	var currentLine strings.Builder
	currentWidth := 0

	// Normalize the text to handle full-width characters
	normalized := width.Narrow.String(line)

	for _, r := range normalized {
		charWidth := runeWidth(r)

		// If adding this character would exceed the line width
		if currentWidth+charWidth > maxWidth && currentWidth > 0 {
			// Check if there's a space in the last 10 runes of the current line
			lineStr := currentLine.String()
			lastSpaceIndex := -1

			// Look for the last space in the last 10 runes
			start := max(0, len(lineStr)-10)
			for j := len(lineStr) - 1; j >= start; j-- {
				if lineStr[j] == ' ' {
					lastSpaceIndex = j
					break
				}
			}

			// If we found a space, wrap there
			if lastSpaceIndex > 0 {
				// Split at the space
				beforeSpace := lineStr[:lastSpaceIndex]
				afterSpace := lineStr[lastSpaceIndex+1:] + string(r)

				lines = append(lines, strings.TrimSpace(beforeSpace))
				currentLine.Reset()
				currentLine.WriteString(afterSpace)
				currentWidth = runeWidth(r) // Reset width for the new line
			} else {
				// No space found, wrap at current position and add hyphen
				lines = append(lines, strings.TrimSpace(currentLine.String())+"-")
				currentLine.Reset()
				currentLine.WriteRune(r)
				currentWidth = charWidth
			}
		} else {
			// Add the character to the current line
			currentLine.WriteRune(r)
			currentWidth += charWidth
		}
	}

	// Add the last line if it has content
	if currentLine.Len() > 0 {
		lines = append(lines, strings.TrimSpace(currentLine.String()))
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
	p.Size(1, 1)   // set font size
	p.Align(escpos.AlignCenter)
	p.PrintLn("Hello Humphrey")

	p.Size(2, 2)
	p.Font(escpos.FontB) // change font
	p.Align(escpos.AlignLeft)

	// Sanitize message to prevent injection
	message = strings.TrimSpace(message)
	if message == "" {
		return fmt.Errorf("message cannot be empty")
	}

	printMessageLines(message)

	p.Align(escpos.AlignCenter)
	p.Barcode(fmt.Sprintf("%d", num), escpos.BarcodeTypeCODE39) // print barcode
	p.Align(escpos.AlignLeft)

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
