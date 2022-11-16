package main

import (
	"os"

	"github.com/gorilla/mux"
	"github.com/robertlestak/sigc/internal/cache"
	"github.com/robertlestak/sigc/internal/server"
	"github.com/robertlestak/sigc/internal/worker"
	log "github.com/sirupsen/logrus"
)

func init() {
	ll, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		ll = log.InfoLevel
	}
	log.SetLevel(ll)
	if err := cache.Init(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	l := log.WithFields(log.Fields{
		"app": "sigc",
		"fn":  "main",
	})
	l.Debug("start")
	port := os.Getenv("PORT")
	var arg string
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}
	if arg == "worker" {
		worker.Start()
	} else {
		if os.Getenv("BACKGROUND_WORKER") == "true" {
			go worker.Start()
		}
		server.Router = mux.NewRouter()
		if err := server.Server(port); err != nil {
			l.Fatal(err)
		}
	}
}
