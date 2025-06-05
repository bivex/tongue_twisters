package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// TongueTwister represents a single tongue twister with its metadata
type TongueTwister struct {
	Number string `json:"number"`
	Date   string `json:"date"`
	Text   string `json:"text"`
}

// PageResult represents the result from scraping a single page
type PageResult struct {
	PageNum   int
	Twisters  []TongueTwister
	Error     error
}

func main() {
	// Parse command line flags
	concurrencyFlag := flag.Int("concurrency", runtime.NumCPU(), "Number of concurrent workers (default: number of CPU cores)")
	outputDirFlag := flag.String("output", "tongue_twisters", "Directory to save output files")
	flag.Parse()

	// Validate concurrency flag
	concurrency := *concurrencyFlag
	if concurrency < 1 {
		concurrency = 1
	} else if concurrency > 20 {
		log.Printf("Warning: High concurrency level (%d) might get you rate limited. Consider using a lower value.", concurrency)
	}

	// Create output directory
	outputDir := *outputDirFlag
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Base URL and total pages (from the HTML: "Всего: 4286 на 215 страницах по 20 на каждой странице")
	baseURL := "https://skorogovorki.my-collection.ru/skorogovorki-cat4"
	totalPages := 215

	// Collect all tongue twisters
	var allTwisters []TongueTwister
	var mutex sync.Mutex // To protect allTwisters from concurrent access
	startTime := time.Now()

	fmt.Printf("Starting to scrape %d pages with %d concurrent workers. This may take a while...\n", 
		totalPages, concurrency)
	
	// Create channels for jobs and results
	jobs := make(chan int, totalPages)
	results := make(chan PageResult, totalPages)
	
	// Launch worker goroutines
	var wg sync.WaitGroup
	for w := 1; w <= concurrency; w++ {
		wg.Add(1)
		go worker(w, baseURL, jobs, results, &wg)
	}
	
	// Send jobs (page numbers) to the workers
	for page := 1; page <= totalPages; page++ {
		jobs <- page
	}
	close(jobs)
	
	// Start a goroutine to collect results
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Map to track completed pages and save results in order
	completed := make(map[int]bool)
	completedCount := 0
	resultsByPage := make(map[int]PageResult)
	
	// Process results as they come in
	for result := range results {
		if result.Error != nil {
			log.Printf("Error scraping page %d: %v", result.PageNum, result.Error)
			continue
		}
		
		// Store result for ordered processing
		resultsByPage[result.PageNum] = result
		
		// Process results in order when possible
		for page := 1; page <= totalPages; page++ {
			if !completed[page] && resultsByPage[page].Twisters != nil {
				pageResult := resultsByPage[page]
				
				// Process the page result
				mutex.Lock()
				for _, twister := range pageResult.Twisters {
					saveToFile(twister, outputDir)
					allTwisters = append(allTwisters, twister)
				}
				mutex.Unlock()
				
				completedCount++
				completed[page] = true
				
				// Calculate and display progress
				progress := float64(completedCount) / float64(totalPages) * 100
				elapsed := time.Since(startTime)
				estimatedTotal := elapsed.Seconds() / (float64(completedCount) / float64(totalPages))
				remaining := time.Duration(estimatedTotal-elapsed.Seconds()) * time.Second
				
				fmt.Printf("[%.1f%%] Completed page %d: found %d tongue twisters (total so far: %d) (Est. remaining: %v)\n", 
					progress, page, len(pageResult.Twisters), len(allTwisters), remaining.Round(time.Second))
				
				// Save progress periodically (every 20 pages)
				if completedCount%20 == 0 {
					mutex.Lock()
					saveAllToJSON(allTwisters, outputDir)
					mutex.Unlock()
					fmt.Printf("Periodic progress saved to JSON after %d pages\n", completedCount)
				}
			} else if !completed[page] {
				// This page hasn't been processed yet, so we need to wait
				break
			}
		}
	}
	
	// Save all tongue twisters to a single JSON file
	saveAllToJSON(allTwisters, outputDir)
	
	elapsed := time.Since(startTime)
	fmt.Printf("Scraping completed! Total tongue twisters: %d (Time elapsed: %s)\n", 
		len(allTwisters), elapsed.Round(time.Second))
}

// worker function that processes jobs from the jobs channel
func worker(id int, baseURL string, jobs <-chan int, results chan<- PageResult, wg *sync.WaitGroup) {
	defer wg.Done()
	
	for page := range jobs {
		// Construct page URL
		pageURL := baseURL
		if page > 1 {
			pageURL = fmt.Sprintf("%s-num%d.html", baseURL, page)
		} else {
			pageURL = baseURL + ".html"
		}
		
		fmt.Printf("Worker %d: Scraping page %d: %s\n", id, page, pageURL)
		
		// Fetch and parse the page with retry mechanism
		var twisters []TongueTwister
		var err error
		maxRetries := 3
		
		for retries := 0; retries < maxRetries; retries++ {
			twisters, err = scrapePageTwisters(pageURL)
			if err == nil {
				break
			}
			log.Printf("Worker %d: Error scraping page %d (attempt %d/%d): %v", id, page, retries+1, maxRetries, err)
			if retries < maxRetries-1 {
				log.Printf("Worker %d: Retrying in 2 seconds...", id)
				time.Sleep(2 * time.Second)
			}
		}
		
		results <- PageResult{
			PageNum:  page,
			Twisters: twisters,
			Error:    err,
		}
		
		// Be nice to the server and add a small delay
		time.Sleep(500 * time.Millisecond)
	}
}

// scrapePageTwisters extracts tongue twisters from a single page
func scrapePageTwisters(url string) ([]TongueTwister, error) {
	// Make HTTP request with proper headers
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var twisters []TongueTwister

	// Find all tongue twister tables
	doc.Find("table.bgcolor4").Each(func(i int, tableSelection *goquery.Selection) {
		var twister TongueTwister

		// Extract number
		numberText := tableSelection.Find("th:first-child small").Text()
		parts := strings.Split(numberText, "№")
		if len(parts) > 1 {
			twister.Number = strings.TrimSpace(parts[1])
		}

		// Extract date
		twister.Date = strings.TrimSpace(tableSelection.Find("th:last-child small").Text())
		
		// Extract text
		twister.Text = strings.TrimSpace(tableSelection.Find("tr.bgcolor1 td").Text())

		if twister.Number != "" && twister.Text != "" {
			twisters = append(twisters, twister)
		}
	})

	return twisters, nil
}

// saveToFile saves a tongue twister to a file in the output directory
func saveToFile(twister TongueTwister, outputDir string) {
	// Create a clean filename
	filename := filepath.Join(outputDir, fmt.Sprintf("twister_%s.txt", twister.Number))
	
	// Create content with metadata
	content := fmt.Sprintf("Number: %s\nDate: %s\n\n%s\n", 
		twister.Number, 
		twister.Date, 
		twister.Text)
	
	// Write to file
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		log.Printf("Error saving file %s: %v", filename, err)
	}
}

// saveAllToJSON saves all tongue twisters to a single JSON file
func saveAllToJSON(twisters []TongueTwister, outputDir string) {
	filename := filepath.Join(outputDir, "all_twisters.json")
	
	// Create JSON data
	jsonData, err := json.MarshalIndent(twisters, "", "  ")
	if err != nil {
		log.Printf("Error creating JSON: %v", err)
		return
	}
	
	// Write to file
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		log.Printf("Error saving JSON file %s: %v", filename, err)
		return
	}
}

// downloadImage downloads an image from a URL and saves it to the output directory
func downloadImage(imageURL, outputDir, filename string) error {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	out, err := os.Create(filepath.Join(outputDir, filename))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
} 