package main

import (
	"context"
	"go-media-server/handlers"

	"github.com/cshum/imagor"
	"github.com/cshum/imagor/loader/httploader"
	"github.com/cshum/imagor/vips"
	"github.com/gin-gonic/gin"
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

	r.GET("/resize", func(dd *gin.Context) {
		handlers.ResizeImageHandler(app, ctx, dd)
	})
	r.GET("/frame", handlers.GetFirstFrameHandler)

	r.Run(":3030")
}
