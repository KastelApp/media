package main

import (
	"context"
	"go-media-server/handlers"

	"github.com/cshum/imagor"
	"github.com/cshum/imagor/loader/httploader"
	"github.com/cshum/imagor/vips"
	"github.com/gin-gonic/gin"
	"regexp"
)

func main() {
	r := gin.Default()

	app := imagor.New(
		imagor.WithLoaders(httploader.New()),
		imagor.WithProcessors(vips.NewProcessor()),
	)

	ctx := context.Background()

	if err := app.Startup(ctx); err != nil {
		panic(err)
	}

	defer app.Shutdown(ctx)

	r.GET("/external/*url", func(c *gin.Context) {
		handlers.ResizeImageHandler(app, ctx, c)
	})
	r.GET("/frame/*url", handlers.GetFirstFrameHandler)
	r.GET("/stream/*url", handlers.StreamVideoHandler)
	r.GET("/thumbhash/*url", handlers.GetThumbHash)

	r.GET("/avatar/:id/:hash", func(c *gin.Context) {
		handlers.HandleAvatar(app, ctx, c)
	})

	r.GET("/file/:channelId/:fileId/:fileName", func(c *gin.Context) {
		handlers.HandleFile(app, ctx, c)
	})

	r.GET("/logo/:logo", func(c *gin.Context) {
		// ? Logo's are stored ./logos. Each wone is like icon-0.png, icon-1.png, etc.
		logo := c.Param("logo")

		if logo == "" {
			c.JSON(400, gin.H{"error": "Invalid logo"})
			
			return
		}

		match := regexp.MustCompile(`^icon-\d+.png$`).MatchString(logo)

		if !match {
			c.JSON(400, gin.H{"error": "Invalid logo"})
			
			return
		}

		c.File("./logos/" + logo)
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the media server!",
		})
	})

	println("Server started on port 3030")

	r.Run(":3030")
}
