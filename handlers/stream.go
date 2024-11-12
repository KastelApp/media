package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func StreamVideoHandler(c *gin.Context) {
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

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		c.JSON(http.StatusInternalServerError, "")

		fmt.Println(err)

		return
	}

	if rangeHeader := c.GetHeader("Range"); rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}

	resp, err := client.Do(req)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, "")

		fmt.Println(err)

		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		c.JSON(http.StatusInternalServerError, "")

		fmt.Println(resp.StatusCode) // ? some people may just block our ips which is fine, sad but fine. Print the status code just in case

		return
	}

	for k, v := range resp.Header {
		for _, vv := range v {
			c.Writer.Header().Add(k, vv)
		}
	}

	c.Status(resp.StatusCode)

	_, err = io.Copy(c.Writer, resp.Body)

	if err != nil {
		c.JSON(http.StatusInternalServerError, "")
	}
}
