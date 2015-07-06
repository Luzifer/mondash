package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/Luzifer/mondash/config"
	"github.com/Luzifer/mondash/storage"
	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"

	_ "github.com/flosch/pongo2-addons"
)

var (
	templates = make(map[string]*pongo2.Template)
	store     storage.Storage
	cfg       *config.Config
)

func main() {
	preloadTemplates()

	var err error
	cfg = config.Load()
	store, err = storage.GetStorage(cfg)
	if err != nil {
		fmt.Printf("An error occurred while loading the storage handler: %s", err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/", handleRedirectWelcome).
		Methods("GET")
	r.HandleFunc("/create", handleCreateRandomDashboard).
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

	http.Handle("/", logHTTPRequest(r))
	http.ListenAndServe(cfg.Listen, nil)
}

func logHTTPRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, r *http.Request) {
		start := time.Now().UnixNano()
		w := NewLogResponseWriter(res)

		h.ServeHTTP(w, r)

		d := (time.Now().UnixNano() - start) / 1000
		log.Printf("%s %s %d %dµs %dB\n", r.Method, r.URL.Path, w.Status, d, w.Size)
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
