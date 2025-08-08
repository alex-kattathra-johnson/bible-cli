package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

const (
	apiBaseURL = "https://api.esv.org/v3/passage/text/"
)

//go:embed verses.json
var versesJSON []byte

type VersesData struct {
	Verses []string `json:"verses"`
}

var bibleVerses []string

func init() {
	var data VersesData
	if err := json.Unmarshal(versesJSON, &data); err != nil {
		panic(fmt.Sprintf("Failed to load verses: %v", err))
	}
	bibleVerses = data.Verses
}

type ESVResponse struct {
	Query       string   `json:"query"`
	Canonical   string   `json:"canonical"`
	Parsed      [][]int  `json:"parsed"`
	Passages    []string `json:"passages"`
	PassageMeta []struct {
		Canonical    string `json:"canonical"`
		ChapterStart []int  `json:"chapter_start"`
		ChapterEnd   []int  `json:"chapter_end"`
		PrevVerse    int    `json:"prev_verse"`
		NextVerse    int    `json:"next_verse"`
	} `json:"passage_meta"`
}

type BibleClient struct {
	apiKey string
	client *http.Client
}

func NewBibleClient(apiKey string) *BibleClient {
	return &BibleClient{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (bc *BibleClient) FetchVerse(reference string) (*ESVResponse, error) {
	params := url.Values{}
	params.Add("q", reference)
	params.Add("include-headings", "false")
	params.Add("include-footnotes", "false")
	params.Add("include-verse-numbers", "false")
	params.Add("include-short-copyright", "false")
	params.Add("include-passage-references", "false")
	params.Add("include-selahs", "false")       // Disable "Selah" notations
	params.Add("include-poetry-lines", "false") // Disable poetry line markers

	fullURL := fmt.Sprintf("%s?%s", apiBaseURL, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+bc.apiKey)

	resp, err := bc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var esvResp ESVResponse
	if err := json.Unmarshal(body, &esvResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &esvResp, nil
}

func (bc *BibleClient) GetRandomVerse() (*ESVResponse, error) {
	randomRef := bibleVerses[rand.Intn(len(bibleVerses))]
	return bc.FetchVerse(randomRef)
}

func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		// Default width if we can't get terminal size
		return 80
	}
	return width
}

func displayVerse(verse *ESVResponse) {
	if verse == nil || len(verse.Passages) == 0 {
		fmt.Println("No passage found")
		return
	}

	// Use the canonical reference from the API response
	reference := verse.Canonical
	passageText := strings.TrimSpace(verse.Passages[0])

	// Get terminal width and calculate box width
	termWidth := getTerminalWidth()
	width := termWidth - 4 // Leave some margin
	if width < 40 {
		width = 40 // Minimum width
	}
	if width > 120 {
		width = 120 // Cap max width for readability
	}

	// Simple border style for better compatibility
	fmt.Println()
	fmt.Println(strings.Repeat("═", width))

	// Center the reference
	refPadding := (width - len(reference)) / 2
	if refPadding < 0 {
		refPadding = 0
	}
	fmt.Printf("%s%s\n", strings.Repeat(" ", refPadding), reference)

	fmt.Println(strings.Repeat("─", width))

	// Word wrap and display the passage text
	lines := strings.Split(passageText, "\n")
	for _, line := range lines {
		wrappedLines := wrapText(line, width-2)
		for _, wrapped := range wrappedLines {
			fmt.Printf(" %s\n", wrapped)
		}
	}

	fmt.Println(strings.Repeat("═", width))
	fmt.Println()
}

func wrapText(text string, maxWidth int) []string {
	if len(text) <= maxWidth {
		return []string{text}
	}

	var result []string
	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		if len(currentLine)+len(word)+1 > maxWidth {
			if currentLine != "" {
				result = append(result, currentLine)
			}
			currentLine = word
		} else {
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		}
	}

	if currentLine != "" {
		result = append(result, currentLine)
	}

	return result
}

func main() {
	apiKey := os.Getenv("ESV_TOKEN")
	if apiKey == "" {
		fmt.Println("Please set the ESV_TOKEN environment variable with your ESV API key.")
		fmt.Println("You can get a free API key at: https://api.esv.org/")
		fmt.Println("\nExample: export ESV_TOKEN='your_api_key_here'")
		os.Exit(1)
	}

	client := NewBibleClient(apiKey)

	if len(os.Args) > 1 {
		reference := strings.Join(os.Args[1:], " ")
		fmt.Printf("Fetching: %s\n", reference)
		verse, err := client.FetchVerse(reference)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		displayVerse(verse)
	} else {
		verse, err := client.GetRandomVerse()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		displayVerse(verse)
	}
}
