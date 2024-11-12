package handlers

import (
	"context"
	"fmt"
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

	if len(url) < 8 {
		c.String(http.StatusBadRequest, "")

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
		c.String(http.StatusBadRequest, "")
		return
	}

	domainParts := strings.Split(domain, "/")
	serviceDomain := c.Request.Host

	if len(domainParts) > 0 {
		domain = domainParts[0]
	}

	if domain == serviceDomain {
		c.String(http.StatusInternalServerError, "")

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
		c.String(http.StatusBadRequest, "")

		return
	}

	if size != 0 && (width != 0 || height != 0) {
		c.String(http.StatusBadRequest, "")

		return
	}

	if size != 0 {
		width = size
		height = size
	}

	if width > 4096 || height > 4096 {
		c.String(http.StatusBadRequest, "The providede width or height is too large")

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
			c.String(http.StatusBadRequest, "Unsupported image format")

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
		c.String(http.StatusInternalServerError, "")
		fmt.Println(err)

		return
	}

	reader, _, err := blob.NewReader()

	if err != nil {
		c.String(http.StatusInternalServerError, "")
		fmt.Println(err)

		return
	}

	defer reader.Close()

	contentType := blob.ContentType()
	c.Header("Content-Type", contentType)
	io.Copy(c.Writer, reader)
}
