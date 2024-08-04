package handlers

import (
	"image"
	_ "image/jpeg" // Import JPEG support
	_ "image/png"  // Import PNG support
	"net/http"
	"strings"
	"sync"
	"time"

	"encoding/base64"

	"github.com/galdor/go-thumbhash"
	"github.com/gin-gonic/gin"
)

var (
	cache           = make(map[string]*cacheEntry)
	cacheMutex      = &sync.Mutex{}
	cacheExpiration = 10 * time.Minute
)

type cacheEntry struct {
	hash      string
	timestamp time.Time
}

func GetThumbHash(c *gin.Context) {
	url := c.Param("url")

	if len(url) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL parameter is required"})
		return
	}

	url = url[1:] // Remove the leading /

	var isHttps bool
	var domain string

	if strings.HasPrefix(url, "https:/") {
		isHttps = true
		domain = url[7:]
		url = url[7:]
	} else if strings.HasPrefix(url, "http:/") {
		isHttps = false
		domain = url[6:]
		url = url[6:]
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
		return
	}

	domainParts := strings.Split(domain, "/")
	if len(domainParts) > 0 {
		domain = domainParts[0]
	}

	serviceDomain := c.Request.Host
	if domain == serviceDomain {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid request to the same domain"})
		return
	}

	if isHttps {
		url = "https://" + url
	} else {
		url = "http://" + url
	}

	// Check cache first
	cacheMutex.Lock()
	if entry, found := cache[url]; found && time.Since(entry.timestamp) < cacheExpiration {
		cacheMutex.Unlock()
		c.JSON(http.StatusOK, gin.H{"thumbhash": entry.hash})
		return
	}
	cacheMutex.Unlock()

	// Fetch the image
	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch image"})
		return
	}
	defer resp.Body.Close()

	// Decode the image
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode image"})
		return
	}

	// Compute the thumbhash
	hash := thumbhash.EncodeImage(img)
	encodedHash := base64.StdEncoding.EncodeToString(hash)

	// Update the cache
	cacheMutex.Lock()
	cache[url] = &cacheEntry{
		hash:      encodedHash,
		timestamp: time.Now(),
	}
	cacheMutex.Unlock()

	// Return the thumbhash
	c.JSON(http.StatusOK, gin.H{"thumbhash": encodedHash})
}
