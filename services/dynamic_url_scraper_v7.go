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
	LCASelector string                        `json:"lcaSelector,omitempty"` // Optional direct LCA selector
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
}

func main() {
    initClickhouse()
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

    if groupFields, ok := selection.Groups["fields"]; ok {
        results, err := findAndExtractData(selection.URL, groupFields, selection.LCASelector)
        if err != nil {
            log.Printf("Error scraping URL %s: %v", selection.URL, err)
            http.Error(w, "Failed to scrape the page", http.StatusInternalServerError)
            return
        }

        jsonData, err := json.Marshal(results)
        if err != nil {
            http.Error(w, "Failed to serialize scraped data", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        w.Write(jsonData)
    } else {
        http.Error(w, "No field group 'fields' found in request", http.StatusBadRequest)
    }
}

func findAndExtractData(url string, fields map[string]string, lcaSelector string) ([]map[string]interface{}, error) {
    doc, err := goquery.NewDocument(url)
    if err != nil {
        log.Printf("Error loading document: %v", err)
        return nil, err
    }

	// lcaSelector, err := findLowestCommonAncestorSelector(doc, fields)
	// if err != nil {
	// 	log.Printf("Error finding LCA selector: %v", err)
	// 	return nil, err
	// }

	// lcaSelector = "div.quote"
	// log.Printf("LCA Selector found: %s", lcaSelector)

    return extractDataFromAncestors(doc, lcaSelector, fields), nil
}

func findLowestCommonAncestorSelector(doc *goquery.Document, fields map[string]string) (string, error) {
    if len(fields) == 0 {
        return "", fmt.Errorf("no fields provided")
    }

    type pathInfo struct {
        path  []*goquery.Selection
        depth int
    }

    // Map to store the paths of all elements
    elementPaths := make(map[string][]pathInfo)

    // Retrieve and store all paths for each field's elements
    for field, selector := range fields {
        elements := doc.Find(selector)
        if elements.Length() == 0 {
            return "", fmt.Errorf("no elements found for selector: %s", selector)
        }

        elements.Each(func(i int, s *goquery.Selection) {
            var path []*goquery.Selection
            for n := s; n.Length() > 0; n = n.Parent() {
                path = append([]*goquery.Selection{n}, path...)
            }
            elementPaths[field] = append(elementPaths[field], pathInfo{path: path, depth: len(path)})
        })
    }

    // Find the lowest common ancestor
    var lca *goquery.Selection
    minDepth := int(^uint(0) >> 1) // Initialize to max int

    // Initialize lca to the root of the first element's path
    for _, paths := range elementPaths {
        if len(paths) > 0 && paths[0].depth < minDepth {
            minDepth = paths[0].depth
            lca = paths[0].path[minDepth-1]
        }
    }

    // Compare all paths to find the common deepest element
    for _, paths := range elementPaths {
        for depth := 0; depth < minDepth; depth++ {
            current := paths[0].path[depth]
            allMatch := true
            for _, path := range paths {
                if path.path[depth].Get(0) != current.Get(0) {
                    allMatch = false
                    break
                }
            }
            if allMatch {
                lca = current
            } else {
                break
            }
        }
    }

    if lca == nil {
        return "", fmt.Errorf("no common ancestor found")
    }

    // Generate a unique selector for the LCA
    tag := goquery.NodeName(lca)
    id, exists := lca.Attr("id")
    if exists && id != "" {
        return tag + "#" + id, nil
    }
    classes, exists := lca.Attr("class")
    if exists && classes != "" {
        classList := strings.Split(classes, " ")
        return tag + "." + strings.Join(classList, "."), nil
    }
    return tag, nil
}


func extractDataFromAncestors2(doc *goquery.Document, lcaSelector string, fields map[string]string) []map[string]interface{} {
    var results []map[string]interface{}
    doc.Find(lcaSelector).Each(func(_ int, s *goquery.Selection) {
        result := make(map[string]interface{})
        for fieldName, selector := range fields {
            if text := s.Find(selector).Text(); text != "" {
                result[fieldName] = strings.TrimSpace(text)
            }
        }
        if len(result) > 0 {
            results = append(results, result)
        }
    })
    return results
}

func extractDataFromAncestors(doc *goquery.Document, lcaSelector string, fields map[string]string) []map[string]interface{} {
    var results []map[string]interface{}
    doc.Find(lcaSelector).Each(func(_ int, s *goquery.Selection) {
        result := make(map[string]interface{})
        for fieldName, fieldSelector := range fields {
            // Find and extract text specifically for each field within the LCA
            fieldValue := s.Find(fieldSelector).Text()
            if fieldValue != "" {
                result[fieldName] = strings.TrimSpace(fieldValue)
            }
        }
        if len(result) > 0 { // Ensure the result map is not empty before appending
            results = append(results, result)
        }
    })
    return results
}
