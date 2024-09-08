package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "github.com/PuerkitoBio/goquery"
    "github.com/gorilla/mux"
)

// FieldSelection stores URL and groups of selectors
type FieldSelection struct {
    URL    string                          `json:"url"`
    Groups map[string]map[string]string    `json:"groups"`
}

// scrapePageForFields processes the webpage and extracts data based on provided selectors
func scrapePageForFields(url string, fields map[string]string) ([]map[string]interface{}, error) {
    var results []map[string]interface{}

    // Fetch the URL
    doc, err := goquery.NewDocument(url)
    if err != nil {
        return nil, fmt.Errorf("failed to load the document: %v", err)
    }

    // Initialize a slice for each key in fields
    fieldData := make(map[string][]string)

    // Collect data for each selector
    for label, selector := range fields {
        doc.Find(selector).Each(func(index int, item *goquery.Selection) {
            fieldData[label] = append(fieldData[label], item.Text())
        })
    }

    // Assuming all fields have the same number of elements
    if len(fieldData) == 0 {
        return results, nil
    }

    // Determine the length of slices in fieldData to sync all fields
    maxLen := 0
    for _, data := range fieldData {
        if len(data) > maxLen {
            maxLen = len(data)
        }
    }

    // Build the results array of maps
    for i := 0; i < maxLen; i++ {
        item := make(map[string]interface{})
        for label, data := range fieldData {
            if i < len(data) {
                item[label] = data[i]
            } else {
                item[label] = "" // Handle missing data points by setting them to an empty string
            }
        }
        results = append(results, item)
    }

    return results, nil
}

// scrapeHandler handles the POST request, extracts data, and returns JSON response
func scrapeHandler(w http.ResponseWriter, r *http.Request) {
    var selection FieldSelection

    // Decode JSON body
    if err := json.NewDecoder(r.Body).Decode(&selection); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Validate URL
    if selection.URL == "" {
        http.Error(w, "URL must be provided", http.StatusBadRequest)
        return
    }

    // Assume only one group for simplicity
    for _, fields := range selection.Groups {
        results, err := scrapePageForFields(selection.URL, fields)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        // Encode results to JSON and send response
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(results)
        return
    }
}

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/scrape", scrapeHandler).Methods("POST")

    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}
