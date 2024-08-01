package handlers

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/cshum/imagor"
	"github.com/cshum/imagor/imagorpath"
	"github.com/gin-gonic/gin"
)

func ResizeImageHandler(app *imagor.Imagor, ctx context.Context, c *gin.Context) {

	// url := c.Query("url")

	// ? url is everything past the /resize/ in the path
	url := c.Param("url")

	println(url);

	// ? we want to remove the first / and https?:/ from the url
	// ? we do want to keep if its http or https tho
	
	url = url[1:]

	isHttps := false

	if url[:5] == "https" {
		isHttps = true
		url = url[7:]
	} else {
		url = url[6:]
	}


	if isHttps {
		url = "https://" + url
	} else {
		url = "http://" + url
	}


	size, _ := strconv.Atoi(c.Query("size"))
	width, _ := strconv.Atoi(c.Query("width"))
	height, _ := strconv.Atoi(c.Query("height"))
	imageType := c.Query("type")

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
		Image:  url,
		Width:  width,
		Height: height,
		Smart:  true,
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
