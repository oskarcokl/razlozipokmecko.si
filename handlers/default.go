package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"text/template"

	m "github.com/oskarcokl/razlozipokmecko.si/models"
	"github.com/oskarcokl/razlozipokmecko.si/services"
	"github.com/oskarcokl/razlozipokmecko.si/tmpl"
)

const EDIT_PATH = "/edit/"
const VIEW_PATH = "/view/"
const SAVE_PATH = "/save/"
const LIST_VIEW_PATH = "/list-view/"


var pattern = filepath.Join("tmpl", "*.html")
var templates = template.Must(template.ParseGlob(pattern))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9-čžš]+)$")


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
	name, err := getName(r); if err != nil {
		http.NotFound(w, r)
	}

    p, err := h.ps.LoadPage(name)
    if err != nil {
        http.Redirect(w, r, EDIT_PATH + name, http.StatusFound)
        return
    }

	// Whos responsibility is rendering templates?
    component := tmpl.ViewPage(p.Name, p.Title, string(p.Body))
    component.Render(context.Background(), w)
}


func (h *DefaultHandler) editHandler(w http.ResponseWriter, r *http.Request) {
	name, err := getName(r); if err != nil {
		http.NotFound(w, r)
	}

    p, err := h.ps.LoadPage(name)
    if err != nil {
        // if page doesn't exists create one
        p = &m.Page{Name: name}
    }
    component := tmpl.EditPage(p.Name, p.Title, string(p.Body))
    component.Render(context.Background(), w)
}


func (h *DefaultHandler) saveHandler(w http.ResponseWriter, r *http.Request) {
	name, err := getName(r); if err != nil {
		http.NotFound(w, r)
	}

    body := r.FormValue("body")
    title := r.FormValue("title")
    p := &m.Page{Name: name, Title: title, Body: []byte(body)}
    p, err = h.ps.SavePage(p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, VIEW_PATH + p.Name, http.StatusFound)
}


func (h *DefaultHandler) listViewHandler(w http.ResponseWriter, r *http.Request) {
    pages := h.ps.LoadAllPages()
    component := tmpl.ListView(pages)
    component.Render(context.Background(), w)
}

func (h *DefaultHandler) renderTemplate(w http.ResponseWriter, tmpl string, p *m.Page) {
    err := templates.ExecuteTemplate(w, tmpl + ".html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func getName(r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		return "", errors.New("path not found")
	}

	return m[2], nil
}