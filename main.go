package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go-media-server/config"
	"go-media-server/handlers"
	"os"

	"regexp"

	"github.com/cshum/imagor"
	"github.com/cshum/imagor/loader/httploader"
	"github.com/cshum/imagor/vips"
	"github.com/gin-gonic/gin"
)

var serverConfig config.Config

func init() {
	configFile := "default"

	if len(os.Args) > 1 && os.Args[1] == "--config" {
		configFile = os.Args[2]
	}

	file, err := os.Open(fmt.Sprintf("configs/%s.json", configFile))

	if err != nil {
		panic("Could not open config file")
	}

	defer file.Close()

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&serverConfig)

	if err != nil {
		panic(err)
	}
}

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
	r.GET("/metadata/*url", handlers.GetMetadata)

	r.GET("/avatar/:id/:hash", func(c *gin.Context) {
		handlers.HandleAvatar(app, ctx, c, serverConfig)
	})

	r.GET("/file/:channelId/:fileId/:fileName", func(c *gin.Context) {
		handlers.HandleFile(app, ctx, c, serverConfig)
	})

	r.GET("/logo/:logo", func(c *gin.Context) {
		// ? Logo's are stored ./logos. Each wone is like icon-0.png, icon-1.png, etc.
		logo := c.Param("logo")

		if logo == "" {
			c.String(400, "")

			return
		}

		match := regexp.MustCompile(`^icon-\d+.png$`).MatchString(logo)

		if !match {
			c.String(400, "")

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
		c.String(200, "Welcome to the media server!")
	})

	println("Server started on port 3030")

	r.Run(":3030")
}
