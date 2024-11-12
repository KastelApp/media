package handlers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	_ "golang.org/x/image/webp"

	"github.com/galdor/go-thumbhash"
	"github.com/gin-gonic/gin"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

var (
	cache           = make(map[string]*cacheEntry)
	cacheMutex      = &sync.Mutex{}
	cacheExpiration = 10 * time.Minute
)

type cacheEntry struct {
	metadata  gin.H
	timestamp time.Time
}

func GetMetadata(c *gin.Context) {
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
		c.String(http.StatusInternalServerError, "Loop back URL")

		return
	}

	if isHttps {
		url = "https://" + url
	} else {
		url = "http://" + url
	}

	cacheMutex.Lock()

	if entry, found := cache[url]; found && time.Since(entry.timestamp) < cacheExpiration {
		cacheMutex.Unlock()
		c.JSON(http.StatusOK, entry.metadata)
		return
	}

	cacheMutex.Unlock()

	resp, err := http.Get(url)

	if err != nil {
		c.String(http.StatusInternalServerError, "")

		fmt.Println(err)

		return
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	var metadata gin.H

	fmt.Println(contentType)

	if strings.HasPrefix(contentType, "image/") {
		img, _, err := image.Decode(resp.Body)

		if err != nil {
			c.String(http.StatusInternalServerError, "")

			fmt.Println(err)

			return
		}

		hash := thumbhash.EncodeImage(img)
		encodedHash := base64.StdEncoding.EncodeToString(hash)

		metadata = gin.H{
			"width":     img.Bounds().Dx(),
			"height":    img.Bounds().Dy(),
			"thumbhash": encodedHash,
		}

	} else if strings.HasPrefix(contentType, "video/") {
		buf := new(bytes.Buffer)
		err := ffmpeg.Input(url).
			Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2"}).
			WithOutput(buf, os.Stdout, os.Stderr).
			Run()

		if err != nil {
			c.String(http.StatusInternalServerError, "")

			fmt.Println(err)

			return
		}

		img, _, err := image.Decode(buf)
		if err != nil {
			c.String(http.StatusInternalServerError, "")

			fmt.Println(err)

			return
		}

		hash := thumbhash.EncodeImage(img)

		encodedHash := base64.StdEncoding.EncodeToString(hash)

		metadata = gin.H{
			"width":     img.Bounds().Dx(),
			"height":    img.Bounds().Dy(),
			"thumbhash": encodedHash,
		}

	} else {
		c.String(http.StatusBadRequest, "Unsupported media type")

		return
	}

	cacheMutex.Lock()
	cache[url] = &cacheEntry{
		metadata:  metadata,
		timestamp: time.Now(),
	}
	cacheMutex.Unlock()

	c.JSON(http.StatusOK, metadata)
}
