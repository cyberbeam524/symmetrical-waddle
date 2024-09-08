package main

import (
    "io"
    "net/http"
    "github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
)

func main() {
    r := gin.Default()
	r.Use(cors.Default())  // Allows all origins

    r.GET("/proxy", func(c *gin.Context) {
        url := c.Query("url")
        if url == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "URL parameter is required"})
            return
        }

        resp, err := http.Get(url)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch webpage"})
            return
        }
        defer resp.Body.Close()

        // Set the content type to whatever was returned by the fetched URL
        contentType := resp.Header.Get("Content-Type")
        c.Header("Content-Type", contentType)

        // Copy the response body directly to the client
        io.Copy(c.Writer, resp.Body)
    })

    r.Run(":8081") // listen and serve on 0.0.0.0:8080
}
