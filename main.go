package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"golang.org/x/exp/slog"
)

var (
	emotextServer           = flag.String("emotext-server", EmotextServer, "emotext server url. Or set EMOTEXT_SERVER env var.")
	musicstoreMurecomServer = flag.String("musicstore-murecom", MusicstoreMurecomServer, "musicstore murecom server url. Or set MUSICSTORE_MURECOM_SERVER env var.")
	listenAddr              = flag.String("listen-addr", "127.0.0.1:8007", "listen address")
)

func main() {
	flag.Parse()

	EmotextServer = *emotextServer
	MusicstoreMurecomServer = *musicstoreMurecomServer

	slog.Info("mureader murecom init.", "EmotextServer", EmotextServer, "MusicstoreMurecomServer", MusicstoreMurecomServer)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	  }))

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"hello": "murecom-gw4reader"})
	})
	r.POST("/murecom", MureaderMurecomHandler)

	srv := &http.Server{
		Addr:    *listenAddr,
		Handler: r,
	}

	run(srv)
}

func run(srv *http.Server) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] srv.ListenAndServe error: %v. Exit...", err)
		}
	}()
	slog.Info("mureader murecom server started.", "listenAddr", srv.Addr)

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	slog.Warn("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	slog.Warn("Server exiting")
}
