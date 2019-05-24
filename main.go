package main

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	httphelper "github.com/Luzifer/go_helpers/http"
	"github.com/Luzifer/mondash/storage"
	"github.com/Luzifer/rconfig"
)

var (
	store storage.Storage
	cfg   = struct {
		APIToken    string `flag:"api-token" env:"API_TOKEN" description:"API Token used for the /welcome dashboard (you can choose your own)"`
		BaseURL     string `flag:"baseurl" env:"BASE_URL" description:"The Base-URL the application is running on for example https://mondash.org"`
		FrontendDir string `flag:"frontend-dir" default:"./frontend" description:"Directory to serve frontend assets from"`
		Storage     string `flag:"storage" default:"file:///data" description:"Storage engine to use"`

		Listen         string `flag:"listen" default:":3000" description:"Address to listen on"`
		LogLevel       string `flag:"log-level" default:"info" description:"Set log level (debug, info, warning, error)"`
		VersionAndExit bool   `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	version = "dev"
)

func init() {
	rconfig.AutoEnv(true)
	if err := rconfig.ParseAndValidate(&cfg); err != nil {
		log.Fatalf("Unable to parse commandline options: %s", err)
	}

	if l, err := log.ParseLevel(cfg.LogLevel); err == nil {
		log.SetLevel(l)
	} else {
		log.Fatalf("Invalid log level: %s", err)
	}

	if cfg.VersionAndExit {
		fmt.Printf("share %s\n", version)
		os.Exit(0)
	}
}

func main() {
	var err error

	if store, err = storage.GetStorage(cfg.Storage); err != nil {
		log.WithError(err).Fatal("Unable to load storage handler")
	}

	r := mux.NewRouter()
	r.Use( // Sort: Outermost to innermost wrapper
		httphelper.NewHTTPLogHandler,
		httphelper.GzipHandler,
		genericHeader,
	)

	r.HandleFunc("/", handleRedirectWelcome).
		Methods(http.MethodGet)
	r.HandleFunc("/app.js", handleAppJS).
		Methods(http.MethodGet)

	r.HandleFunc("/create", handleCreateRandomDashboard).
		Methods(http.MethodGet)
	r.HandleFunc("/{dashid}.json", handleDisplayDashboardJSON).
		Methods(http.MethodGet)
	r.HandleFunc("/{dashid}", handleDisplayDashboard).
		Methods(http.MethodGet)

	r.HandleFunc("/{dashid}/{metricid}", handlePutMetric).
		Methods(http.MethodPut)

	r.HandleFunc("/{dashid}", handleDeleteDashboard).
		Methods(http.MethodDelete)
	r.HandleFunc("/{dashid}/{metricid}", handleDeleteMetric).
		Methods(http.MethodDelete)

	go runWelcomePage()

	if err := http.ListenAndServe(cfg.Listen, r); err != nil {
		log.WithError(err).Fatal("HTTP server ended unexpectedly")
	}
}

func genericHeader(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, r *http.Request) {
		res.Header().Set("X-Application-Version", version)
		h.ServeHTTP(res, r)
	})
}

func generateAPIKey() string {
	t := time.Now().String()
	sum := md5.Sum([]byte(t))
	return fmt.Sprintf("%x", sum)
}
