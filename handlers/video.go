package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
)

func extractFirstFrame(videoURL, outputFile string) error {
	cmd := exec.Command(
		"ffmpeg",
		"-i", videoURL,
		"-vf", "thumbnail",
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
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL parameter is required"})
		return
	}

	outputFile := "first_frame.png"

	defer os.Remove(outputFile)

	if err := extractFirstFrame(url, outputFile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract frame", "err": err.Error()})
		return
	}

	file, err := os.Open(outputFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open extracted frame", "err": err.Error()})
		return
	}
	defer file.Close()

	c.Header("Content-Type", "image/png")
	if _, err := io.Copy(c.Writer, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send image", "err": err.Error()})
	}
}
