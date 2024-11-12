package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	ffmpeg "github.com/u2takey/ffmpeg-go"
	"github.com/gin-gonic/gin"
)

func extractFirstFrame(videoURL string, width, height int) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)

	err := ffmpeg.Input(videoURL).
		Filter("thumbnail", ffmpeg.Args{}).
		Filter("scale", ffmpeg.Args{fmt.Sprintf("%d:%d", width, height)}).
		Output("pipe:", ffmpeg.KwArgs{"frames:v": "1", "format": "image2"}).
		WithOutput(buf, nil).
		Run()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg error: %w", err)
	}

	return buf, nil
}

func GetFirstFrameHandler(c *gin.Context) {
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
		c.String(http.StatusBadRequest, "Invalid URL")
		return
	}

	domainParts := strings.Split(domain, "/")
	serviceDomain := c.Request.Host

	if len(domainParts) > 0 {
		domain = domainParts[0]
	}

	if domain == serviceDomain {
		c.String(http.StatusInternalServerError, "Loop back URL")

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

	if size != 0 && (width != 0 || height != 0) {
		c.String(http.StatusBadRequest, "size and width/height cannot be used together")

		return
	}

	if size != 0 {
		width = size
		height = size
	}

	if width > 1024 || height > 1024 {
		c.String(http.StatusBadRequest, "width and height cannot be greater than 1024")

		return
	}

	if size == 0 && width == 0 && height == 0 {
		width = 854
		height = 480
	}

	frameBuf, err := extractFirstFrame(url, width, height)
	
	if err != nil {
		c.String(http.StatusInternalServerError, "")

		fmt.Println(err)

		return
	}

	c.Header("Content-Type", "image/png")
	
	if _, err := io.Copy(c.Writer, frameBuf); err != nil {
		c.String(http.StatusInternalServerError, "") // ? shouldn't happen
	}
}
