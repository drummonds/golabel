package main

import (
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

var p *escpos.Printer
var tmpl *template.Template

const maxLineLength = 27

// min function that was missing
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max function for smart wrapping
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// wrapTextUnicode is a Unicode-aware text wrapping function
func wrapTextUnicode(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{text}
	}

	var lines []string
	var currentLine strings.Builder
	currentWidth := 0

	// Normalize the text to handle full-width characters
	normalized := width.Narrow.String(text)

	for _, r := range normalized {
		charWidth := runeWidth(r)

		// If adding this character would exceed the line width
		if currentWidth+charWidth > maxWidth && currentWidth > 0 {
			// Start a new line
			lines = append(lines, strings.TrimSpace(currentLine.String()))
			currentLine.Reset()
			currentWidth = 0
		}

		// Add the character to the current line
		currentLine.WriteRune(r)
		currentWidth += charWidth
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

// wrapText is a more efficient text wrapping function
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)

	if len(words) == 0 {
		return []string{""}
	}

	currentLine := words[0]

	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
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

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>GoLabel - TM-T20III Label Printer</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 600px;
            margin: 50px auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            text-align: center;
            margin-bottom: 30px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 5px;
            font-weight: bold;
            color: #555;
        }
        input[type="text"], input[type="number"], textarea {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 5px;
            font-size: 16px;
            box-sizing: border-box;
        }
        textarea {
            resize: vertical;
            min-height: 100px;
            font-family: inherit;
        }
        button {
            background-color: #007bff;
            color: white;
            padding: 12px 30px;
            border: none;
            border-radius: 5px;
            font-size: 16px;
            cursor: pointer;
            width: 100%;
        }
        button:hover {
            background-color: #0056b3;
        }
        .status {
            margin-top: 20px;
            padding: 10px;
            border-radius: 5px;
            text-align: center;
        }
        .success {
            background-color: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        .error {
            background-color: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }
        .footer {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid #eee;
            text-align: center;
            color: #666;
            font-size: 12px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>GoLabel - TM-T20III Printer Control</h1>
        <form method="POST" action="/print">
            <div class="form-group">
                <label for="message">Message to Print:</label>
                <textarea id="message" name="message" placeholder="Enter your message here..." required></textarea>
            </div>
            <div class="form-group">
                <label for="barcode">Barcode Number:</label>
                <input type="number" id="barcode" name="barcode" value="5" min="1" max="999999">
            </div>
            <button type="submit">Print Label</button>
        </form>
        {{if .Status}}
        <div class="status {{if .Success}}success{{else}}error{{end}}">
            {{.Status}}
        </div>
        {{end}}
        <div class="footer">
            <div>Version: {{.Version}}</div>
            <div>Built: {{.BuildDate}}</div>
        </div>
    </div>
</body>
</html>
`

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
	var err error

	// Initialize template once
	tmpl, err = template.New("printer").Parse(htmlTemplate)
	if err != nil {
		fmt.Println("Error parsing template:", err)
		return
	}

	p, err = escpos.NewUSBPrinterByPath("") // empty string will do a self discovery
	if err != nil {
		fmt.Println("Error initializing printer:", err)
		return
	}

	fmt.Printf("Starting GoLabel web server on http://localhost:8080\n")
	fmt.Printf("Version: %s\n", version.GetVersionInfo())
	fmt.Println("Printer initialized successfully")

	http.HandleFunc("/", handlePrint)
	http.HandleFunc("/print", handlePrint)

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
