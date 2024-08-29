package handlers

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/cshum/imagor"
	"github.com/gin-gonic/gin"
	"mime"
)

/*
HandleFile like HandleAvatar intercepts the request and routes it to the ResizeImageHandler if the mime type is an image
*/
func HandleFile(app *imagor.Imagor, ctx context.Context, c *gin.Context) {
	cdnUrl := "http://localhost:9000/files"

	channelId := c.Param("channelId")
	fileId := c.Param("fileId")
	fileName := c.Param("fileName")

	fullUrl := cdnUrl + "/" + channelId + "/" + fileId + "/" + fileName

	ext := strings.ToLower(filepath.Ext(fileName))
	mimeType := mime.TypeByExtension(ext)

	if mimeType != "" && mimeType != "image/gif" && strings.HasPrefix(mimeType, "image") {
		fixedUrl := "/http:/localhost:9000/files/" + channelId + "/" + fileId + "/" + fileName
		c.Params = append(c.Params, gin.Param{Key: "url", Value: fixedUrl})
		ResizeImageHandler(app, ctx, c)
	} else {
		resp, err := http.Get(fullUrl)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch the file"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			c.JSON(resp.StatusCode, gin.H{"error": "File not found"})
			return
		}

		c.Header("Content-Type", mimeType)
		c.Status(http.StatusOK)
		_, err = io.Copy(c.Writer, resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stream the file"})
		}
	}
}
