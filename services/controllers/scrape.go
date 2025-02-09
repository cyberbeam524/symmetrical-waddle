package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"github.com/PuerkitoBio/goquery"
	"github.com/playwright-community/playwright-go"
)

func ScrapeHandler(w http.ResponseWriter, r *http.Request) {
    var selection FieldSelection
    if err := json.NewDecoder(r.Body).Decode(&selection); err != nil {
        log.Println("Invalid request body")
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

	// Validate the `pages` field and set a default if it's missing or invalid
	if selection.Pages == nil {
		log.Println("Pages not provided. Defaulting to 1 page.")
		defaultPages := 1
		selection.Pages = &defaultPages
	} else if *selection.Pages <= 0 {
		log.Println("Invalid pages value provided. Defaulting to 1 page.")
		defaultPages := 1
		selection.Pages = &defaultPages
	}

	log.Printf("Scraping up to %d pages", *selection.Pages)
    pw, err := playwright.Run()
    if err != nil {
        log.Fatalf("could not start playwright: %v", err)
    }
    browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
        Headless: playwright.Bool(false), // set to false in development
    })
    if err != nil {
        log.Fatalf("could not launch browser: %v", err)
    }
    // page, err := browser.NewPage()
    context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	})
	if err != nil {
		log.Fatalf("could not create browser context: %v", err)
	}

	page, err := context.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}


    if err != nil {
        log.Fatalf("could not create page: %v", err)
    }

    // pass a playwright.Page and the script will be injected
    // err = stealth.Inject(page)
    if err != nil {
        log.Fatalf("could not inject stealth script: %v", err)
    }
    // Load the requested URL
    page.Goto(selection.URL)

    // Scroll through the page to load all the content
    scrollingScript := `
        // scroll down the page 10 times
        const scrolls = 10
        let scrollCount = 0

        // scroll down and then wait for 0.5s
        const scrollInterval = setInterval(() => {
            window.scrollTo(0, document.body.scrollHeight)
            scrollCount++

            if (scrollCount === scrolls) {
                clearInterval(scrollInterval)
            }
        }, 500)
    `
    // execute the custom JavaScript script on the page
    _, err = page.Evaluate(scrollingScript, []interface{}{})
    if err != nil {
        log.Fatal("Could not perform the JS scrolling logic:", err)
    }
	// Create a map to track existing selectors for quick lookup
	existingSelectors := make(map[string]struct{})
	for _, selector := range blockerSelectors {
		existingSelectors[selector] = struct{}{}
	}
	// Split and process the additional blockers
	if selection.Blockers != "" {
		additionalBlockers := strings.Split(selection.Blockers, ",")
		// Add only new selectors that aren't already in blockerSelectors
		for _, blocker := range additionalBlockers {
			blocker = strings.TrimSpace(blocker)
			if _, exists := existingSelectors[blocker]; !exists && blocker != "" {
				blockerSelectors = append(blockerSelectors, blocker)
				existingSelectors[blocker] = struct{}{}
			}
		}
	}
	if err := handleBlockers(page, blockerSelectors); err != nil {
		log.Printf("Error handling blockers: %v", err)
	}
	// log.Fatal("Pausing here:")
	// page.Pause();
	log.Println("Scrolling done")
	log.Println("Scraping pages numer: %s", selection.Pages);
    results := make([]map[string]interface{}, 0)
	currPage := 0;
    kafkaTopic := selection.KafkaTopic;
    for {
        groupFields, ok := selection.Groups["fields"]
        if !ok {
            log.Println("No field group 'fields' found in request")
            http.Error(w, "No field group 'fields' found in request", http.StatusBadRequest)
            return
        }
        results = append(results, extractDataFromAncestors3(page, selection.LCASelector, groupFields)...)
		log.Println("results: %s", results);
        // Check if there's a "next page" button and click it
		log.Println("PaginationSelector: %s", selection.PaginationSelector);
		nextButton := page.Locator(selection.PaginationSelector)
		count, _ := nextButton.Count()

		if count == 0 {
			log.Println("No more pages to load, stopping pagination.")
			break
		}
		// Handle blockers after the initial page load
		if err := handleBlockers(page, blockerSelectors); err != nil {
			log.Printf("Error handling blockers: %v", err)
		}

		// Take a screenshot of the classified listing
		screenshotPath := fmt.Sprintf("page-%d.png", currPage+1)
		if _, err = page.Screenshot(playwright.PageScreenshotOptions{
			Path: playwright.String(screenshotPath),
		}); err != nil {
			log.Fatalf("could not create screenshot: %v", err)}

		fmt.Println("Screenshot of listing taken and saved to", screenshotPath)
		
		log.Println("Clicking next page button...")
		currentURL := page.URL()
		if err := nextButton.Click(); err != nil {
			log.Printf("Failed to click next page button: %v", err)
			break
		}
		// Wait for the page to load after clicking
		log.Println("Waiting for the next page to load...")
		page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{})
		newURL := page.URL()
		if currentURL == newURL {
			log.Println("Page URL did not change after clicking next page. Stopping pagination.")
			break
		}

		log.Printf("Navigated to new page: %s", newURL)

		currPage += 1;
		if currPage >= *selection.Pages{
			log.Printf("Completed all pagese: %s -- %s", currPage, selection.Pages)
			break;
		}
        // Stream scraped data in batch
		if err := produceBatchToKafka(kafkaTopic, results); err != nil {
			http.Error(w, "Failed to stream to Kafka", http.StatusInternalServerError)
			return
		}
        results = results[:0]
        log.Printf("Scraped data streamed to Kafka topic: %s for page: %s", kafkaTopic, currPage)
    }

    // Clean up the Playwright resources
    page.Close()
    browser.Close()
    pw.Stop()

    if len(results) > 0 {
        // log.Printf("Scraped data streamed to Kafka topic: %s", kafkaTopic)
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Scraping and streaming completed"))


    } else {
        log.Println("func scrapeHandler failed")
        http.Error(w, "No data found on the page", http.StatusNotFound)
    }
}

func scrapeHandler2(w http.ResponseWriter, r *http.Request) {
    log.Println("func scrapeHandler executing");
    var selection FieldSelection
    if err := json.NewDecoder(r.Body).Decode(&selection); err != nil {
        log.Println("Invalid request body");
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
        log.Println(results);
        jsonData, err := json.Marshal(results)
        if err != nil {
            log.Println("Failed to serialize scraped data");
            http.Error(w, "Failed to serialize scraped data", http.StatusInternalServerError)
            return
        }

        log.Println("func scrapeHandler success");
        w.Header().Set("Content-Type", "application/json")
        w.Write(jsonData)
    } else {
        log.Println("func scrapeHandler failed");
        http.Error(w, "No field group 'fields' found in request", http.StatusBadRequest)
    }
}


type FieldSelection struct {
    URL    string                         `json:"url"`
    Groups map[string]map[string]string   `json:"groups"` // Map of groups with their field selectors
	LCASelector string                        `json:"lcaSelector,omitempty"` // Optional direct LCA selector
	PaginationSelector string                         `json:"paginationSelector"`
	Blockers string                         `json:"blockers"`
	Pages *int                           `json:"pages,omitempty"`
    KafkaTopic string
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

    return extractDataFromAncestors2(doc, lcaSelector, fields), nil
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


var blockerSelectors = []string{
    "button.accept-cookies",       // Example cookie consent button
    "div.modal-close",             // Modal close button
    "button.close-popup",          // Generic popup close button
    ".overlay-dismiss",            // Overlay dismiss button
    ".interstitial-dismiss",       // Ad or interstitial close button
	// "#onetrust-accept-btn-handler", // Button with the specific ID
}

func handleBlockers(page playwright.Page, blockerSelectors []string) error {
    for _, selector := range blockerSelectors {
        log.Printf("Checking for blocker: %s", selector)
        element := page.Locator(selector)
        count, _ := element.Count()

        if count > 0 {
			var timeout float64
			timeout = 1000 
            log.Printf("Blocker found: %s. Attempting to click...", selector)
            err := element.Click(playwright.LocatorClickOptions{
				Timeout: &timeout,
			})
            if err != nil {
                log.Printf("Failed to click blocker: %s. Error: %v", selector, err)
                continue
            }
            log.Printf("Successfully clicked blocker: %s", selector)
        }
    }
    return nil
}

func extractDataFromAncestors3(page playwright.Page, lcaSelector string, fields map[string]string) []map[string]interface{} {
    var results []map[string]interface{}

	log.Println("lcaSelector: %s", lcaSelector);
    elements, err := page.Locator(lcaSelector).All()
	if err != nil {
        log.Fatalf("Could not get the product node: %v", err)
    }
    for _, element := range elements {
        result := make(map[string]interface{})
        for fieldName, selector := range fields {
			// log.Printf("Processing selector for field '%s': %s", fieldName, selector)

            // Check if the element exists and is visible
            fieldElement := element.Locator(selector)
            isVisible, _ := fieldElement.IsVisible()
            if !isVisible {
                // log.Printf("Field '%s' is not visible for selector: %s", fieldName, selector)
                continue
            }

             // Get the text content of the field
            text, err := fieldElement.TextContent()
            if err != nil {
                // log.Printf("Error getting text content for field '%s': %v", fieldName, err)
                continue
            }
            if text == "" {
                // log.Printf("Field '%s' has empty text for selector: %s", fieldName, selector)
                continue
            }

            result[fieldName] = strings.TrimSpace(text)
            // log.Printf("Extracted value for field '%s': %s", fieldName, result[fieldName])

        }
        if len(result) > 0 {
            results = append(results, result)
        }
    }
    // page.Pause()
    return results
}

func flattenJSON(input map[string]interface{}) map[string]interface{} {
	flattened := make(map[string]interface{})
	var flatten func(map[string]interface{}, string)

	flatten = func(currentMap map[string]interface{}, prefix string) {
		for key, value := range currentMap {
			fullKey := key
			if prefix != "" {
				fullKey = prefix + "." + key
			}

			switch v := value.(type) {
			case map[string]interface{}:
				flatten(v, fullKey)
			default:
				flattened[fullKey] = v
			}
		}
	}

	flatten(input, "")
	return flattened
}