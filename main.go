package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"log"

	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"

	"github.com/flosch/pongo2"
	"github.com/go-martini/martini"

	_ "github.com/flosch/pongo2-addons"
)

var templates = make(map[string]*pongo2.Template)
var s3Storage *s3.Bucket

func main() {
	preloadTemplates()

	// Initialize S3 storage
	awsAuth, err := aws.EnvAuth()
	if err != nil {
		log.Fatal(err)
	}
	s3Conn := s3.New(awsAuth, aws.USEast)
	s3Storage = s3Conn.Bucket(os.Getenv("S3Bucket"))

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

	go runWelcomePage()

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
