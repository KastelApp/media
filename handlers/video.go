package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func extractFirstFrame(videoURL, outputFile string, width, height int) error {
	cmd := exec.Command(
		"ffmpeg",
		"-i", videoURL,
		"-vf", fmt.Sprintf("thumbnail,scale=%d:%d", width, height),
		"-frames:v", "1",
		"-f", "image2",
		outputFile,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg error: %w, output: %s", err, output)
	}
	return nil
}

func GetFirstFrameHandler(c *gin.Context) {

	url := c.Param("url")

	// ? we want to remove the first / and https?:/ from the url
	// ? we do want to keep if its http or https tho

	if len(url) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL parameter is required"})

		return
	}

	url = url[1:]

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

	size, _ := strconv.Atoi(c.Query("size"))
	width, _ := strconv.Atoi(c.Query("width"))
	height, _ := strconv.Atoi(c.Query("height"))

	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL parameter is required"})
		return
	}

	if size != 0 && (width != 0 || height != 0) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "size and width/height cannot be used together"})
		return
	}

	if size != 0 {
		width = size
		height = size
	}

	if width > 1024 || height > 1024 { // ? no reason for that high of a resolution
		c.JSON(http.StatusBadRequest, gin.H{"error": "width and height cannot be greater than 1024"})

		return
	}

	if size == 0 && width == 0 && height == 0 {
		// ? default to a 480p image
		width = 854
		height = 480
	}

	outputFile := "first_frame.png"

	defer os.Remove(outputFile)

	if err := extractFirstFrame(url, outputFile, width, height); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract frame"})
		return
	}

	file, err := os.Open(outputFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open extracted frame"})
		return
	}
	defer file.Close()

	c.Header("Content-Type", "image/png")
	if _, err := io.Copy(c.Writer, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send image"})
	}
}
