package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	m "github.com/oskarcokl/razlozipokmecko.si/models"
	"github.com/oskarcokl/razlozipokmecko.si/services"
)

const EDIT_PATH = "/edit/"
const VIEW_PATH = "/view/"
const SAVE_PATH = "/save/"
const LIST_VIEW_PATH = "/list-view/"


var pattern = filepath.Join("tmpl", "*.html")
var templates = template.Must(template.ParseGlob(pattern))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")


type DefaultHandler struct {
	ps *services.PageService
}


func New(ps *services.PageService) *DefaultHandler {
	return &DefaultHandler{
		ps: ps,
	}
}


func (h *DefaultHandler) ServeHTTP() {
	// Define handlers
    http.HandleFunc(VIEW_PATH, h.viewHandler)
    http.HandleFunc(EDIT_PATH, h.editHandler)
    http.HandleFunc(SAVE_PATH, h.saveHandler)
    http.HandleFunc(LIST_VIEW_PATH, h.listViewHandler)

    fmt.Println("Server running on port 8000")
    log.Fatal(http.ListenAndServe(":8000", nil))
}


func (h *DefaultHandler) viewHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(r); if err != nil {
		http.NotFound(w, r)
	}

    p, err := h.ps.LoadPage(strings.Join(strings.Split(title, "-"), " "))
    if err != nil {
        http.Redirect(w, r, EDIT_PATH + title, http.StatusFound)
        return
    }

	// Whos responsibility is rendering templates?
    h.renderTemplate(w, "view", p)
}


func (h *DefaultHandler) editHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(r); if err != nil {
		http.NotFound(w, r)
	}


    p, err := h.ps.LoadPage(title)
    if err != nil {
        // if page doesn't exists create one
        p = &m.Page{Title: title}
    }
    p.Title = strings.Join(strings.Split(p.Title, "-"), " ")
    h.renderTemplate(w, "edit", p)
}


func (h *DefaultHandler) saveHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(r); if err != nil {
		http.NotFound(w, r)
	}

    body := r.FormValue("body")
    p := &m.Page{Title: title, Body: []byte(body)}
    err = h.ps.SavePage(p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, VIEW_PATH + title, http.StatusFound)
}


func (h *DefaultHandler) listViewHandler(w http.ResponseWriter, r *http.Request) {
    pages := h.ps.LoadAllPages()

    err := templates.ExecuteTemplate(w, "list-view.html", pages); if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func (h *DefaultHandler) renderTemplate(w http.ResponseWriter, tmpl string, p *m.Page) {
    err := templates.ExecuteTemplate(w, tmpl + ".html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func getTitle(r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		return "", errors.New("path not found")
	}

	return m[2], nil
}