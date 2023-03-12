package main

import (
	services "chat_room"
	"context"
	"flag"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
)

var addr = flag.String("addr", "localhost:8080", "this demo's address")

func main() {
	go services.ListenAndServe()
	h := server.Default(config.Option{F: func(o *config.Options) {
		o.Addr = *addr
	}})
	h.GET("/ws", func(_ context.Context, c *app.RequestContext) {
		c.Set("name", c.Query("name"))
		services.ServeWs(c)
	})
	h.Spin()
}
