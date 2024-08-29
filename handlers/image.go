package handlers

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/cshum/imagor"
	"github.com/cshum/imagor/imagorpath"
	"github.com/gin-gonic/gin"
)

func ResizeImageHandler(app *imagor.Imagor, ctx context.Context, c *gin.Context) {
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
	imageType := c.Query("format")

	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL query param is required"})
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

	if width > 4096 || height > 4096 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "width and height cannot be greater than 4096"})

		return
	}

	if imageType != "" {
		possibleTypes := []string{"jpeg", "png", "webp", "gif"}

		validType := false

		for _, t := range possibleTypes {
			if t == imageType {
				validType = true
			}
		}

		if !validType {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image type"})
			return
		}
	}

	filters := []imagorpath.Filter{}

	if imageType != "" {
		filters = append(filters, imagorpath.Filter{Name: "format", Args: imageType})
	}

	blob, err := app.Serve(ctx, imagorpath.Params{
		Image:   url,
		Width:   width,
		Height:  height,
		Smart:   true,
		Filters: filters,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	reader, _, err := blob.NewReader()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}
	defer reader.Close()

	// ? get the content type of the image
	contentType := blob.ContentType()

	// ? set the content type of the image
	c.Header("Content-Type", contentType)

	io.Copy(c.Writer, reader)
}
