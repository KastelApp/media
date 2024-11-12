package handlers

import (
	"context"
	"fmt"
	"go-media-server/config"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"mime"

	"github.com/cshum/imagor"
	"github.com/gin-gonic/gin"
)

/*
HandleFile like HandleAvatar intercepts the request and routes it to the ResizeImageHandler if the mime type is an image
*/
func HandleFile(app *imagor.Imagor, ctx context.Context, c *gin.Context, serverConfig config.Config) {
	channelId := c.Param("channelId")
	fileId := c.Param("fileId")
	fileName := c.Param("fileName")

	fullUrl := fmt.Sprintf("%s://%s/files/%s/%s/%s", config.GetProtocol(serverConfig.Secure), serverConfig.CdnUrl, channelId, fileId, fileName)

	ext := strings.ToLower(filepath.Ext(fileName))
	mimeType := mime.TypeByExtension(ext)

	if mimeType != "" && mimeType != "image/gif" && strings.HasPrefix(mimeType, "image") {
		fixedUrl := fmt.Sprintf("/%s:/%s/%s/%s/%s", config.GetProtocol(serverConfig.Secure), serverConfig.CdnUrl, channelId, fileId, fileName)
		c.Params = append(c.Params, gin.Param{Key: "url", Value: fixedUrl})
		ResizeImageHandler(app, ctx, c)

	} else {
		resp, err := http.Get(fullUrl)

		if err != nil {
			c.String(http.StatusInternalServerError, "")

			return
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			c.String(resp.StatusCode, "")

			return
		}

		c.Header("Content-Type", mimeType)
		c.Status(http.StatusOK)

		_, err = io.Copy(c.Writer, resp.Body)

		if err != nil {
			c.String(http.StatusInternalServerError, "")

			fmt.Println(err)
		}
	}
}
