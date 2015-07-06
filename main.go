package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Luzifer/mondash/config"
	"github.com/Luzifer/mondash/storage"
	"github.com/flosch/pongo2"
	"github.com/go-martini/martini"

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

	m := martini.Classic()

	// Assets are in assets folder
	m.Use(martini.Static("assets", martini.StaticOptions{Prefix: "/assets"}))

	// Real handlers
	m.Get("/", handleRedirectWelcome)
	m.Get("/create", handleCreateRandomDashboard)
	m.Get("/:dashid", handleDisplayDashboard)

	m.Put("/:dashid/:metricid", handlePutMetric)

	m.Delete("/:dashid", handleDeleteDashboard)
	m.Delete("/:dashid/:metricid", handleDeleteMetric)

	go runWelcomePage(cfg)

	// GO!
	m.Run()
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
