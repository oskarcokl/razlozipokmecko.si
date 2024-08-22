package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"

	m "github.com/oskarcokl/razlozipokmecko.si/models"
	"github.com/oskarcokl/razlozipokmecko.si/services"
	"github.com/oskarcokl/razlozipokmecko.si/tmpl"
)


const EDIT_PATH = "/edit/"
const VIEW_PATH = "/view/"
const SAVE_PATH = "/save/"
const LIST_VIEW_PATH = "/list-view/"


var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9-čžš]+)$")


type DefaultHandler struct {
	es *services.ExplanationService
}


func New(es *services.ExplanationService) *DefaultHandler {
	return &DefaultHandler{
		es: es,
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

    e, err := h.es.LoadExplanation(name)
    if err != nil {
        http.Redirect(w, r, EDIT_PATH + name, http.StatusFound)
        return
    }

	// Whos responsibility is rendering templates?
    component := tmpl.ViewExplanation(e.Name, e.Title, string(e.Body))
    component.Render(context.Background(), w)
}


func (h *DefaultHandler) editHandler(w http.ResponseWriter, r *http.Request) {
	name, err := getName(r); if err != nil {
		http.NotFound(w, r)
	}

    e, err := h.es.LoadExplanation(name)
    if err != nil {
        // if page doesn't exists create one
        e = &m.Explanation{Name: name}
    }
    component := tmpl.EditExplanation(e.Name, e.Title, string(e.Body))
    component.Render(context.Background(), w)
}


func (h *DefaultHandler) saveHandler(w http.ResponseWriter, r *http.Request) {
	name, err := getName(r); if err != nil {
		http.NotFound(w, r)
	}

    body := r.FormValue("body")
    title := r.FormValue("title")
    e := &m.Explanation{Name: name, Title: title, Body: []byte(body)}
    e, err = h.es.SaveExplanation(e)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, VIEW_PATH + e.Name, http.StatusFound)
}


func (h *DefaultHandler) listViewHandler(w http.ResponseWriter, r *http.Request) {
    explanations := h.es.LoadAllExplanations()
    component := tmpl.ListView(explanations)
    component.Render(context.Background(), w)
}


func getName(r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		return "", errors.New("path not found")
	}

	return m[2], nil
}