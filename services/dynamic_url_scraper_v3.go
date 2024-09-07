package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strings"
	"time"

    "github.com/PuerkitoBio/goquery"
    "github.com/gorilla/mux"
    _ "github.com/ClickHouse/clickhouse-go"
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

func fetchSelectableFieldsHandler(w http.ResponseWriter, r *http.Request) {
    url := r.URL.Query().Get("url")
    if url == "" {
        http.Error(w, "URL parameter is missing", http.StatusBadRequest)
        return
    }

    doc, err := goquery.NewDocument(url)
    if err != nil {
        http.Error(w, "Error fetching webpage", http.StatusInternalServerError)
        return
    }

    selectors := make(map[string]string)
    doc.Find("*").Each(func(i int, s *goquery.Selection) {
        if id, exists := s.Attr("id"); exists {
            selectors["id_"+id] = "#" + id
        }
        if class, exists := s.Attr("class"); exists {
            for _, cl := range strings.Split(class, " ") {
                if cl != "" {
                    selectors["class_"+cl] = "." + cl
                }
            }
        }
    })

    response, _ := json.Marshal(selectors)
    w.Header().Set("Content-Type", "application/json")
    w.Write(response)
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

    if err := insertDataIntoClickhouse(selection.URL, string(jsonData)); err != nil {
        log.Printf("Error inserting data into ClickHouse: %v", err)
        http.Error(w, "Failed to insert data into database", http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Scraping and insertion completed successfully")
}

func scrapePageForFields(url string, fields map[string]string) ([]map[string]interface{}, error) {
    doc, err := goquery.NewDocument(url)
    if err != nil {
        return nil, err
    }

    var results []map[string]interface{}
    doc.Find(fields["container"]).Each(func(i int, s *goquery.Selection) {
        product := make(map[string]interface{})
        for fieldName, selector := range fields {
            if fieldName != "container" {
                product[fieldName] = strings.TrimSpace(s.Find(selector).Text())
            }
        }
        results = append(results, product)
    })

    return results, nil
}

func insertDataIntoClickhouse(url, jsonData string) error {
    tx, err := clickhouseConn.Begin()
    if err != nil {
        return err
    }

    stmt, err := tx.Prepare("INSERT INTO scraped_data (url, content, scraped_at) VALUES (?, ?, ?)")
    if err != nil {
        tx.Rollback()
        return err
    }
    defer stmt.Close()

    _, err = stmt.Exec(url, jsonData, time.Now())
    if err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit()
}

func main() {
    initClickhouse()
    r := mux.NewRouter()
    r.HandleFunc("/fetch-selectable-fields", fetchSelectableFieldsHandler).Methods("GET")
    r.HandleFunc("/scrape", scrapeHandler).Methods("POST")
    log.Println("HTTP server started on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}
