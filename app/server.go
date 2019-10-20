package app

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/rs/xid"
)

type Server struct {
	SymptomRepo SymptomRepo
	EventRepo   EventRepo
	Router      *mux.Router
	Env         map[string]string
}

func (s *Server) Routes() {
	// Define router, static files, and middleware
	s.Router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	// Routes to handle requests from browsers and HTML forms
	s.Router.HandleFunc("/", s.handleIndex()).Methods("GET")
	s.Router.HandleFunc("/symptoms", s.handleSymptoms()).Methods("GET")
	s.Router.HandleFunc("/symptoms/add", s.handleAddSymptom()).Methods("GET", "POST")
	s.Router.HandleFunc("/symptoms/remove", s.handleRemoveSymptom()).Methods("POST")
	s.Router.HandleFunc("/symptoms/{id:[0-9]+}", s.handleViewSymptom()).Methods("GET", "POST")
	s.Router.HandleFunc("/symptoms/report", s.handleReportSymptoms()).Methods("GET")
}

/**
 * HTTP handler for get requests to view the homepage
 */
func (s *Server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := "index"
		title := "Home"

		// Handle 404 routes
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			name = "404"
			title = "Uh oh"
		}

		s.getTemplate(name, nil).Execute(w, map[string]interface{}{
			"title": title,
			"env":   s.Env,
		})
	}
}

/**
 * HTTP handler for get requests to view all symptoms
 */
func (s *Server) handleSymptoms() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		symptoms, err := s.SymptomRepo.All()
		if err != nil {
			fmt.Println("Unable to get symptoms")
		}

		s.getTemplate("symptoms/index", nil).Execute(w, map[string]interface{}{
			"title":          "Add Symptom",
			"env":            s.Env,
			csrf.TemplateTag: csrf.TemplateField(r),
			"symptoms":       symptoms,
		})
	}
}

/**
 * HTTP handler for get/post requests to add a symptom to the stacks
 */
func (s *Server) handleAddSymptom() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle save symptom on post
		if r.Method == http.MethodPost {
			// Create new symptom from form values
			b, err := s.SymptomRepo.New(
				r.FormValue("title"),
				r.FormValue("author"),
				r.FormValue("description"),
			)
			if err != nil {
				fmt.Printf("Cannot create symptom: %s", err)
				http.Redirect(w, r, "/symptoms", http.StatusSeeOther)
			}

			if err = s.EventRepo.SymptomAdded(b.ID); err != nil {
				// TODO: Handle event error
			}

			http.Redirect(w, r, "/symptoms", http.StatusSeeOther)
		}

		// Handle save symptom form on get
		s.getTemplate("symptoms/add", nil).Execute(w, map[string]interface{}{
			"title":          "Add Symptom",
			"env":            s.Env,
			csrf.TemplateTag: csrf.TemplateField(r),
		})
	}
}

/**
 * HTTP handler for post requests to remove a symptom from the stacks
 */
func (s *Server) handleRemoveSymptom() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			fmt.Printf("Cannot get ID from form: %s", err)
			http.Redirect(w, r, "/symptoms", http.StatusSeeOther)
		}

		err = s.SymptomRepo.Delete(id)
		if err != nil {
			fmt.Printf("Cannot delete symptom with ID %d: %s", id, err)
			http.Redirect(w, r, "/symptoms", http.StatusSeeOther)
		}

		if err = s.EventRepo.SymptomRemoved(id); err != nil {
			// TODO: Handle event error
		}

		http.Redirect(w, r, "/symptoms", http.StatusSeeOther)
	}
}

/**
 * HTTP handler for get/post requests to view a symptom
 */
func (s *Server) handleViewSymptom() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Redirect(w, r, "/symptoms", http.StatusMovedPermanently)
		}

		b, err := s.SymptomRepo.One(id)
		if err != nil {
			fmt.Printf("Cannot find symptom %d\n", id)
			http.Redirect(w, r, "/symptoms", http.StatusSeeOther)
		}

		es, err := s.EventRepo.AllForSymptom(id)
		if err != nil {
			fmt.Printf("Cannot find symptom %d\n", id)
			http.Redirect(w, r, "/symptoms", http.StatusSeeOther)
		}

		// Handle save symptom on post
		if r.Method == http.MethodPost {
			// Update symptom from form values
			b.Title = r.FormValue("title")
			b.Author = r.FormValue("author")
			b.Description = r.FormValue("description")
			if err = s.SymptomRepo.Update(b); err != nil {
				fmt.Printf("Cannot update symptom: %s", err)
				http.Redirect(w, r, "/symptoms", http.StatusSeeOther)
			}

			if err = s.EventRepo.SymptomAdded(b.ID); err != nil {
				// TODO: Handle event error
			}

			http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		}

		// Handle save symptom form on get
		s.getTemplate("symptoms/_symptom", nil).Execute(w, map[string]interface{}{
			"title":          "Add Symptom",
			"env":            s.Env,
			csrf.TemplateTag: csrf.TemplateField(r),
			"symptom":        b,
			"events":         es,
		})
	}
}

/**
 * HTTP handler for get requests to view status report of symptoms
 */
func (s *Server) handleReportSymptoms() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		symptoms, err := s.SymptomRepo.All()
		if err != nil {
			fmt.Println("Unable to get symptoms")
		}

		s.getTemplate("symptoms/report", nil).Execute(w, map[string]interface{}{
			"title":          "Add Symptom",
			"env":            s.Env,
			csrf.TemplateTag: csrf.TemplateField(r),
			"symptoms":       symptoms,
		})
	}
}

/**
 * Helper function to load all required layouts and partials to load template
 */
func (s *Server) getTemplate(name string, fm template.FuncMap) *template.Template {
	funcMap := template.FuncMap{
		"now": func() int {
			return time.Now().Year()
		},
		"uniqueID": func() string {
			return xid.New().String()
		},
	}
	// Merge custom funcMap
	for k, v := range fm {
		funcMap[k] = v
	}

	t, err := template.New("main.html").Funcs(funcMap).ParseFiles(
		"templates/_layouts/main.html",
		"templates/_meta/data.html",
		"templates/_meta/favicons.html",
		fmt.Sprintf("templates/%s.html", name),
	)
	if err != nil {
		fmt.Printf("Unable to load template %s: \n", name, err)
	}

	return t
}
