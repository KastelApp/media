package handlers

import (
	"context"
	"fmt"
	"go-media-server/config"

	"github.com/cshum/imagor"
	"github.com/gin-gonic/gin"
)

/*
HandleAvatar intercepts the request and routes it to the ResizeImageHandler from the CDN (minio / s3) instance
*/
func HandleAvatar(app *imagor.Imagor, ctx context.Context, c *gin.Context, serverConfig config.Config) {
	id := c.Param("id")
	hash := c.Param("hash")

	fullUrl := fmt.Sprintf("%s://%s/avatars/%s/%s", config.GetProtocol(serverConfig.Secure), serverConfig.CdnUrl, id, hash)

	c.Params = append(c.Params, gin.Param{Key: "url", Value: fullUrl})

	ResizeImageHandler(app, ctx, c)
}
