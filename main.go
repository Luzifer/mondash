package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
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
	s3Conn := s3.New(awsAuth, aws.EUWest)
	s3Storage = s3Conn.Bucket(os.Getenv("S3Bucket"))

	m := martini.Classic()

	// Assets are in assets folder
	m.Use(martini.Static("assets", martini.StaticOptions{Prefix: "/assets"}))

	// Real handlers
	m.Get("/", func(res http.ResponseWriter, req *http.Request) {
		http.Redirect(res, req, "/welcome", 302)
	})

	m.Get("/create", func(res http.ResponseWriter, req *http.Request) {
		urlProposal := generateAPIKey()[0:20]
		_, err := s3Storage.Get(urlProposal)
		for err == nil {
			urlProposal = generateAPIKey()[0:20]
			_, err = s3Storage.Get(urlProposal)
		}
		http.Redirect(res, req, fmt.Sprintf("/%s", urlProposal), http.StatusTemporaryRedirect)
	})

	m.Get("/:dashid", func(params martini.Params, res http.ResponseWriter) {
		dash, err := LoadDashboard(params["dashid"])
		if err != nil {
			dash = &Dashboard{APIKey: generateAPIKey(), Metrics: DashboardMetrics{}}
		}

		// Filter out expired metrics
		metrics := DashboardMetrics{}
		for _, m := range dash.Metrics {
			if m.Meta.LastUpdate.After(time.Now().Add(time.Duration(m.Expires*-1) * time.Second)) {
				metrics = append(metrics, m)
			}
		}

		sort.Sort(sort.Reverse(DashboardMetrics(metrics)))
		renderTemplate("dashboard.html", pongo2.Context{
			"dashid":  params["dashid"],
			"metrics": metrics,
			"apikey":  dash.APIKey,
			"baseurl": os.Getenv("BASE_URL"),
		}, res)
	})

	m.Delete("/:dashid", func(params martini.Params, req *http.Request, res http.ResponseWriter) {
		dash, err := LoadDashboard(params["dashid"])
		if err != nil {
			http.Error(res, "This dashboard does not exist.", http.StatusInternalServerError)
			return
		}

		if dash.APIKey != req.Header.Get("Authorization") {
			http.Error(res, "APIKey did not match.", http.StatusUnauthorized)
			return
		}

		s3Storage.Del(params["dashid"])
		http.Error(res, "OK", http.StatusOK)
	})

	m.Put("/:dashid/:metricid", func(params martini.Params, req *http.Request, res http.ResponseWriter) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(res, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		metricUpdate := NewDashboardMetric()
		err = json.Unmarshal(body, metricUpdate)
		if err != nil {
			http.Error(res, "Unable to unmarshal json", http.StatusInternalServerError)
			return
		}

		dash, err := LoadDashboard(params["dashid"])
		if err != nil {
			if len(req.Header.Get("Authorization")) < 10 {
				http.Error(res, "APIKey is too insecure", http.StatusUnauthorized)
				return
			}
			dash = &Dashboard{APIKey: req.Header.Get("Authorization"), Metrics: DashboardMetrics{}, DashboardID: params["dashid"]}
		}

		if dash.APIKey != req.Header.Get("Authorization") {
			http.Error(res, "APIKey did not match.", http.StatusUnauthorized)
			return
		}

		valid, reason := metricUpdate.IsValid()
		if !valid {
			http.Error(res, fmt.Sprintf("Invalid data: %s", reason), http.StatusInternalServerError)
			return
		}

		updated := false
		for _, m := range dash.Metrics {
			if m.MetricID == params["metricid"] {
				m.Update(metricUpdate)
				updated = true
				break
			}
		}

		if !updated {
			tmp := NewDashboardMetric()
			tmp.MetricID = params["metricid"]
			tmp.Update(metricUpdate)
			dash.Metrics = append(dash.Metrics, tmp)
		}

		dash.Save()

		http.Error(res, "OK", http.StatusOK)
	})

	m.Delete("/:dashid/:metricid", func(params martini.Params, req *http.Request, res http.ResponseWriter) {
		dash, err := LoadDashboard(params["dashid"])
		if err != nil {
			dash = &Dashboard{APIKey: req.Header.Get("Authorization"), Metrics: DashboardMetrics{}, DashboardID: params["dashid"]}
		}

		if dash.APIKey != req.Header.Get("Authorization") {
			http.Error(res, "APIKey did not match.", http.StatusUnauthorized)
			return
		}

		tmp := DashboardMetrics{}
		for _, m := range dash.Metrics {
			if m.MetricID != params["metricid"] {
				tmp = append(tmp, m)
			}
		}
		dash.Metrics = tmp
		dash.Save()

		http.Error(res, "OK", http.StatusOK)
	})

	go RunWelcomePage()

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
		tpl.ExecuteWriter(context, res)
	} else {
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(fmt.Sprintf("Template %s not found!", templateName)))
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
