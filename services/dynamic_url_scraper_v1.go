package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
	"encoding/json" // Add this import

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
	_ "github.com/ClickHouse/clickhouse-go"
)

var (
	clickhouseConn *sql.DB
	cache          = make(map[string]string) // Cache to store the latest data
	cacheMutex     sync.RWMutex              // Mutex to handle concurrent access to cache
)

type FieldSelection struct {
	URL     string   `json:"url"`
	Fields  []string `json:"fields"`
	PageNum int      `json:"page_num"` // Number of pages to scrape
}

// Initialize ClickHouse connection
func initClickhouse() {
	var err error
	// clickhouseConn, err = sql.Open("clickhouse", "tcp://127.0.0.1:9000?debug=true")
	clickhouseConn, err = sql.Open("clickhouse", "tcp://localhost:9000?debug=true")
	if err != nil {
		log.Fatalf("Error connecting to ClickHouse: %v", err)
	}

	// Ensure connection is valid
	if err = clickhouseConn.Ping(); err != nil {
		log.Fatalf("Could not ping ClickHouse: %v", err)
	}

	// Create a table to store the scraped data
	query := `CREATE TABLE IF NOT EXISTS scraped_data (
		url String,
		field String,
		content String,
		scraped_at DateTime
	) ENGINE = MergeTree() ORDER BY scraped_at`

	if _, err := clickhouseConn.Exec(query); err != nil {
		log.Fatalf("Error creating ClickHouse table: %v", err)
	}
}

// API to return available branches of data (mocked)
func branchesHandler(w http.ResponseWriter, r *http.Request) {
	// This is a static mock. Ideally, you should fetch real branches or sections from the target website
	branches := []string{
		"/news",
		"/products",
		"/services",
		"/blog",
	}
	for _, branch := range branches {
		fmt.Fprintf(w, "Branch: %s\n", branch)
	}
}

// API to allow users to select fields from a page
func selectFieldsHandler(w http.ResponseWriter, r *http.Request) {
	// Extract URL from query params
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "Missing URL parameter", http.StatusBadRequest)
		return
	}

	// Fetch and parse the webpage
	doc, err := goquery.NewDocument(url)
	if err != nil {
		http.Error(w, "Error fetching webpage", http.StatusInternalServerError)
		return
	}

	// Extract potential fields (CSS classes, IDs) from the page
	var fields []string
	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		if class, exists := s.Attr("class"); exists && class != "" {
			fields = append(fields, "."+class)
		}
		if id, exists := s.Attr("id"); exists && id != "" {
			fields = append(fields, "#"+id)
		}
	})

	// Send fields as response
	fmt.Fprintf(w, "Selectable Fields from %s:\n", url)
	for _, field := range fields {
		fmt.Fprintf(w, "%s\n", field)
	}
}

// Function to scrape the selected fields from multiple pages
func scrapeSelectedFieldsHandler(w http.ResponseWriter, r *http.Request) {
	// Example request body format:
	// {
	//   "url": "https://example.com",
	//   "fields": [".product-title", ".product-price"],
	//   "page_num": 5
	// }

	// Parse the request body (containing URL, selected fields, number of pages)
	var selection FieldSelection
	if err := json.NewDecoder(r.Body).Decode(&selection); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// For each page, scrape the selected fields
	for i := 1; i <= selection.PageNum; i++ {
		pageURL := fmt.Sprintf("%s?page=%d", selection.URL, i)
		content, err := scrapePageForFields(pageURL, selection.Fields)
		if err != nil {
			log.Printf("Error scraping page %s: %v", pageURL, err)
			continue
		}

		

		// Insert the scraped data into ClickHouse
		for field, data := range content {
			fmt.Printf("%s:%s\n", field, data)
			if err := insertDataIntoClickhouse(selection.URL, field, data); err != nil {
				log.Printf("Error inserting data into ClickHouse: %v", err)
			}
		}
	}

	fmt.Fprintf(w, "Scraping and insertion completed for %s\n", selection.URL)
}

// Scrape a single page for the selected fields
func scrapePageForFields(url string, fields []string) (map[string]string, error) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, field := range fields {
		// Extract content from the selected field (CSS class/ID)
		content := strings.TrimSpace(doc.Find(field).Text())
		if content != "" {
			result[field] = content
		}
	}

	return result, nil
}

// Insert scraped data into ClickHouse
func insertDataIntoClickhouse2(url, field, content string) error {
	query := "INSERT INTO scraped_data (url, field, content, scraped_at) VALUES (?, ?, ?, ?)"
	_, err := clickhouseConn.Exec(query, url, field, content, time.Now())
	return err
}


// Insert scraped data into ClickHouse
func insertDataIntoClickhouse(url, field, content string) error {
    // Start a transaction
    tx, err := clickhouseConn.Begin()
    if err != nil {
        log.Printf("Error starting transaction: %v", err)
        return err
    }

    // Prepare the insert statement within the transaction
    stmt, err := tx.Prepare("INSERT INTO scraped_data (url, field, content, scraped_at) VALUES (?, ?, ?, ?)")
    if err != nil {
        log.Printf("Error preparing statement: %v", err)
        return err
    }
    defer stmt.Close()  // Ensure the statement is closed after the function exits

    // Execute the insert statement
    _, err = stmt.Exec(url, field, content, time.Now())
    if err != nil {
        // Rollback the transaction if the insert fails
        if rbErr := tx.Rollback(); rbErr != nil {
            log.Printf("Error rolling back transaction: %v", rbErr)
            return rbErr
        }
        log.Printf("Error executing insert: %v", err)
        return err
    }

    // Commit the transaction
    if err = tx.Commit(); err != nil {
        log.Printf("Error committing transaction: %v", err)
        return err
    }

    return nil
}


func main() {
	// Initialize ClickHouse connection
	initClickhouse()

	// Create a new router
	r := mux.NewRouter()

	// Define API routes
	r.HandleFunc("/branches", branchesHandler).Methods("GET")
	r.HandleFunc("/select-fields", selectFieldsHandler).Methods("GET")
	r.HandleFunc("/scrape-selected", scrapeSelectedFieldsHandler).Methods("POST")

	// Start the HTTP server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
