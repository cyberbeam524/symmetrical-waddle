package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strings"

    "github.com/PuerkitoBio/goquery"
    "github.com/gorilla/mux"
    _ "github.com/ClickHouse/clickhouse-go"
    "github.com/gorilla/handlers"
)

var (
    clickhouseConn *sql.DB
)

type FieldSelection struct {
    URL    string                         `json:"url"`
    Groups map[string]map[string]string   `json:"groups"` // Map of groups with their field selectors
}

func initClickhouse() {
    var err error
    clickhouseConn, err = sql.Open("clickhouse", "tcp://localhost:9000?debug=true")
    if err != nil {
        log.Fatalf("Error connecting to ClickHouse: %v", err)
    }

    if err = clickhouseConn.Ping(); err != nil {
        log.Fatalf("Could not ping ClickHouse: %v", err)
    }

    query := `CREATE TABLE IF NOT EXISTS scraped_data (
        url String,
        content String,
        scraped_at DateTime
    ) ENGINE = MergeTree() ORDER BY scraped_at`

    if _, err := clickhouseConn.Exec(query); err != nil {
        log.Fatalf("Error creating ClickHouse table: %v", err)
    }
}

func main() {
    // initClickhouse()
    r := mux.NewRouter()
    r.HandleFunc("/scrape", scrapeHandler).Methods("POST")

    corsObj := handlers.AllowedOrigins([]string{"*"})
    headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
    methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

    log.Println("HTTP server started on :8080")
    http.ListenAndServe(":8080", handlers.CORS(corsObj, headersOk, methodsOk)(r))
}

func scrapeHandler(w http.ResponseWriter, r *http.Request) {
    var selection FieldSelection
    if err := json.NewDecoder(r.Body).Decode(&selection); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    results := make(map[string]interface{})
    for groupName, fields := range selection.Groups {
        scrapedData, err := scrapePageForFields(selection.URL, fields)
        if err != nil {
            log.Printf("Error scraping URL %s: %v", selection.URL, err)
            http.Error(w, "Failed to scrape the page", http.StatusInternalServerError)
            continue
        }
        results[groupName] = scrapedData
    }

    jsonData, err := json.Marshal(results)
    if err != nil {
        http.Error(w, "Failed to serialize scraped data", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(jsonData)
}

func scrapePageForFields(url string, fields map[string]string) ([]map[string]interface{}, error) {
    doc, err := goquery.NewDocument(url)
    if err != nil {
        log.Printf("Error loading document: %v", err)
        return nil, err
    }

    var results []map[string]interface{}
    for _, selector := range fields { // Modified to ignore the unused fieldName
        elements := doc.Find(selector)
        elements.Each(func(_ int, element *goquery.Selection) {
            ancestor := findCommonAncestor(element)
            data := extractData(ancestor, fields)
            if len(data) > 0 {
                results = append(results, data)
            }
        })
    }

    if len(results) == 0 {
        return nil, fmt.Errorf("no data extracted with provided selectors")
    }

    return results, nil
}

func findCommonAncestor(element *goquery.Selection) *goquery.Selection {
    return element.Parents().First() // Simplified common ancestor discovery for illustration
}

func extractData(ancestor *goquery.Selection, fields map[string]string) map[string]interface{} {
    data := make(map[string]interface{})
    for field, selector := range fields {
        text := strings.TrimSpace(ancestor.Find(selector).Text())
        if text != "" {
            data[field] = text
        }
    }
    return data
}
