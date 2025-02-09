package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/playwright-community/playwright-go"
)

// Post represents the scraped post data
type Post struct {
	Title string
	Link  string
}

func main() {
	// Initialize Playwright
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("Could not start Playwright: %v", err)
	}

	// Launch Chromium browser
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true), // Set to false for debugging
	})
	if err != nil {
		log.Fatalf("Could not launch browser: %v", err)
	}
	defer browser.Close()

	// Open a new page
	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("Could not create page: %v", err)
	}

	// Navigate to the subreddit page
	fmt.Println("Navigating to Reddit...")
	if _, err := page.Goto("https://www.reddit.com/r/programming/"); err != nil {
		log.Fatalf("Could not navigate to Reddit: %v", err)
	}

	// Wait for the posts to load
	fmt.Println("Waiting for posts to load...")
	_, err = page.WaitForSelector("a.absolute.inset-0", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(20000), // Timeout in milliseconds
	})
	if err != nil {
		log.Fatalf("Posts did not load: %v", err)
	}

	// Scrape the post data
	fmt.Println("Scraping posts...")
	posts := []Post{}
	postElements, err := page.QuerySelectorAll("a.absolute.inset-0")
	if err != nil {
		log.Fatalf("Could not find post elements: %v", err)
	}

	for _, postElement := range postElements {
		// Extract post link
		link, err := postElement.GetAttribute("href")
		if err != nil || link == "" {
			log.Println("Could not extract link:", err)
			continue
		}

		// Extract post title (if available)
		titleElement, err := postElement.QuerySelector("div > h3") // Adjust the selector as needed
		if err != nil || titleElement == nil {
			log.Println("Could not find title element:", err)
			continue
		}

		title, err := titleElement.TextContent()
		if err != nil || strings.TrimSpace(title) == "" {
			log.Println("Could not extract title:", err)
			continue
		}

		// Store the data
		posts = append(posts, Post{
			Title: strings.TrimSpace(title),
			Link:  "https://www.reddit.com" + link,
		})
	}

	// Print the scraped data
	fmt.Println("Scraped Posts:")
	for _, post := range posts {
		fmt.Printf("- Title: %s\n  Link: %s\n", post.Title, post.Link)
	}
}
