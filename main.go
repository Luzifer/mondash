package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	httphelper "github.com/Luzifer/go_helpers/http"
	"github.com/Luzifer/mondash/config"
	"github.com/Luzifer/mondash/storage"

	_ "github.com/Luzifer/mondash/filters"
)

var (
	templates = make(map[string]*pongo2.Template)
	store     storage.Storage
	cfg       *config.Config

	version string
)

func main() {
	preloadTemplates()

	var err error

	if cfg, err = config.Load(); err != nil {
		log.WithError(err).Fatal("Unable to load config")
	}

	if store, err = storage.GetStorage(cfg); err != nil {
		log.WithError(err).Fatal("Unable to load storage handler")
	}

	r := mux.NewRouter()
	r.Use( // "Applied in order they are specified"
		genericHeader,
		httphelper.GzipHandler,
		httphelper.NewHTTPLogHandler,
	)

	r.HandleFunc("/", handleRedirectWelcome).
		Methods("GET")
	r.HandleFunc("/create", handleCreateRandomDashboard).
		Methods("GET")
	r.HandleFunc("/{dashid}.json", handleDisplayDashboardJSON).
		Methods("GET")
	r.HandleFunc("/{dashid}", handleDisplayDashboard).
		Methods("GET")

	r.HandleFunc("/{dashid}/{metricid}", handlePutMetric).
		Methods("PUT")

	r.HandleFunc("/{dashid}", handleDeleteDashboard).
		Methods("DELETE")
	r.HandleFunc("/{dashid}/{metricid}", handleDeleteMetric).
		Methods("DELETE")

	go runWelcomePage(cfg)

	if err := http.ListenAndServe(cfg.Listen, r); err != nil {
		log.WithError(err).Fatal("HTTP server ended unexpectedly")
	}
}

func genericHeader(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, r *http.Request) {
		res.Header().Set("X-Application-Version", version)
	})
}

func generateAPIKey() string {
	t := time.Now().String()
	sum := md5.Sum([]byte(t))
	return fmt.Sprintf("%x", sum)
}

func renderTemplate(templateName string, context pongo2.Context, res http.ResponseWriter) {
	if tpl, ok := templates[templateName]; ok {
		_ = tpl.ExecuteWriter(context, res)
	} else {
		res.WriteHeader(http.StatusInternalServerError)
		_, _ = res.Write([]byte(fmt.Sprintf("Template %s not found!", templateName)))
	}
}

func preloadTemplates() {
	templateNames, err := ioutil.ReadDir("templates")
	if err != nil {
		panic("Templates directory not available!")
	}
	for _, tplname := range templateNames {
		templates[tplname.Name()] = pongo2.Must(pongo2.FromFile(fmt.Sprintf("templates/%s", tplname.Name())))
	}
}
