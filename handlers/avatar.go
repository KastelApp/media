package handlers

import (
	"context"

	"github.com/cshum/imagor"
	"github.com/gin-gonic/gin"
)

/*
HandleAvatar intercepts the request and routes it to the ResizeImageHandler from the CDN (minio / s3) instance
*/
func HandleAvatar(app *imagor.Imagor, ctx context.Context, c *gin.Context) {
	cdnUrl := "/http:/localhost:9000/avatars"

	id := c.Param("id")
	hash := c.Param("hash")

	fullUrl := cdnUrl + "/" + id + "/" + hash

	c.Params = append(c.Params, gin.Param{Key: "url", Value: fullUrl})

	ResizeImageHandler(app, ctx, c)
}
