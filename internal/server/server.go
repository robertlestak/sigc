package server

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/robertlestak/sigc/pkg/client"
	"github.com/robertlestak/sigc/pkg/schema"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

var Router *mux.Router

func healthHandler(w http.ResponseWriter, r *http.Request) {
	l := log.WithFields(log.Fields{
		"app": "server",
		"fn":  "healthHandler",
	})
	l.Debug("start")
	w.WriteHeader(http.StatusOK)
}

func HandleExec(w http.ResponseWriter, r *http.Request) {
	l := log.WithFields(log.Fields{
		"app": "server",
		"fn":  "HandleExec",
	})
	l.Debug("start")
	defer r.Body.Close()
	sr := &schema.SignedRequest{}
	err := json.NewDecoder(r.Body).Decode(sr)
	if err != nil {
		l.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	res, err := client.ExecSignedRequest(sr)
	if err != nil {
		l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		l.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func HandleCreateSignedRequest(w http.ResponseWriter, r *http.Request) {
	l := log.WithFields(log.Fields{
		"app": "server",
		"fn":  "HandleCreateSignedRequest",
	})
	l.Debug("start")
	defer r.Body.Close()
	sr := &schema.SignRequest{}
	err := json.NewDecoder(r.Body).Decode(sr)
	if err != nil {
		l.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = sr.Validate()
	if err != nil {
		l.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	signedRequest, err := sr.CreateSignedRequest()
	if err != nil {
		l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(signedRequest); err != nil {
		l.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func StartServer(port string, corsList []string) error {
	l := log.WithFields(log.Fields{
		"action": "StartServer",
	})
	l.Debug("StartServer")
	if len(corsList) == 0 {
		corsList = []string{"*"}
	}
	c := cors.New(cors.Options{
		AllowedOrigins:   corsList,
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            os.Getenv("CORS_DEBUG") == "true",
	})
	h := c.Handler(Router)
	return http.ListenAndServe(":"+port, h)
}

func Server(port string) error {
	l := log.WithFields(log.Fields{
		"app": "server",
		"fn":  "Server",
	})
	l.Debug("start")
	if os.Getenv("SIGN_SERVER") == "true" {
		Router.HandleFunc("/sign", HandleCreateSignedRequest)
	}
	Router.HandleFunc("/exec", HandleExec)
	Router.HandleFunc("/health", healthHandler)
	if port == "" {
		port = "8080"
	}
	return StartServer(port, []string{"*"})
}
