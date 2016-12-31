package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/99designs/basicauth-go"
	"github.com/facebookgo/httpdown"
	"github.com/pressly/chi"

	"github.com/btobolaski/ejrnl"
	"github.com/btobolaski/ejrnl/workflows"
)

type Config struct {
	Port               int
	Username, Password string
}

type Server struct {
	port        int
	driver      ejrnl.Driver
	server      httpdown.Server
	credentials map[string][]string
}

func New(driver ejrnl.Driver, config Config) (*Server, error) {
	creds := map[string][]string{config.Username: []string{config.Password}}
	return &Server{port: config.Port, driver: driver, credentials: creds}, nil
}

func (s *Server) Start() error {
	router := chi.NewRouter()
	router.Use(basicauth.New("ejrnl", s.credentials))
	router.Get("/", s.index)

	router.Get("/autosize.js", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(autosize))
	})

	router.Route("/entries", func(r chi.Router) {
		r.Get("/new", s.newForm)
		r.Post("/new", s.create)
		r.Get("/:entryId/", s.read)
	})

	options := httpdown.HTTP{
		StopTimeout: 5 * time.Second,
		KillTimeout: 10 * time.Second,
	}

	httpServerOptions := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: router,
	}

	server, err := options.ListenAndServe(httpServerOptions)
	if err != nil {
		return err
	}

	s.server = server
	return nil
}

type entryListing struct {
	Id   string
	Date time.Time
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	listing, index, err := workflows.Listing(s.driver)
	if err != nil {
		log.Printf("Failed to generate listing because %s", err)
		http.Error(w, "A Server Error occured", 500)
		return
	}
	entries := make([]entryListing, len(index))
	for i, val := range index {
		entries[i] = entryListing{
			Id:   listing[val],
			Date: val,
		}
	}
	templateData := struct {
		Title   string
		Entries []entryListing
	}{"Entries", entries}

	err = indexPage.Execute(w, templateData)
	if err != nil {
		log.Printf("Failed to generate listing because %s", err)
		http.Error(w, "A Server Error occured", 500)
	}
}

func (s *Server) newForm(w http.ResponseWriter, r *http.Request) {
	date := time.Now()
	entry := ejrnl.Entry{Date: &date}
	renderForm(workflows.Format(entry), "New Entry", w)
}

func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	entry, err := workflows.Read([]byte(r.Form["text"][0]))
	if err != nil {
		log.Printf("Failed to parse input data because %s", err)
		renderForm(r.Form["text"][0], "Re-edit", w)
		return
	}

	err = s.driver.Write(entry)
	if err != nil {
		log.Printf("Failed to save because %s", err)
		renderForm(r.Form["text"][0], "Re-edit", w)
		return
	}

	http.Redirect(w, r, "/", 303)
}

func (s *Server) read(w http.ResponseWriter, r *http.Request) {
	entryId := chi.URLParam(r, "entryId")
	entry, err := s.driver.Read(entryId)
	if err != nil {
		log.Printf("Error while trying to look up entry %s, %s", entryId, err)
		http.Error(w, "Couldn't find entry with that id", 404)
		return
	}
	renderForm(workflows.Format(entry), fmt.Sprintf("Edit %s", entryId), w)
}

func renderForm(entry, title string, w http.ResponseWriter) {
	data := struct {
		Title, Target, Text string
	}{"New Entry", "/entries/new", entry}
	err := formPage.Execute(w, data)
	if err != nil {
		log.Printf("Failed to generate form because %s", err)
		http.Error(w, "A Server Error occured", 500)
	}
}
