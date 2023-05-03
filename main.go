package main

import (
	"laplace/core"
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/acme/autocert"

)

const HOST = "share.zig.show"

func main() {
	e := echo.New()
	e.Pre(middleware.HTTPSNonWWWRedirect())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit("4K"))
	e.Use(middleware.Gzip())

	e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(HOST)
	e.AutoTLSManager.Cache = autocert.DirCache(".cache")
	e.HideBanner = true

	laplaceServer := core.GetHttp()
	
	e.Any("/*", echo.WrapHandler(laplaceServer))

	println("Listening to ports 80 and 443")
	go e.StartAutoTLS(HOST + ":" + "443")
	e80 := echo.New()
	e80.Pre(middleware.HTTPSNonWWWRedirect())
	e80.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))
	e80.Use(middleware.Recover())
	e80.Use(middleware.BodyLimit("4K"))
	e80.HideBanner = true
	go e80.Start(HOST + ":" + "80")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	e.Shutdown(ctx)
	println("Main server shutdown!")
	if e80 != nil {
		println("Port 80 redirect shutdown!")
		e80.Shutdown(ctx)
	}

}

