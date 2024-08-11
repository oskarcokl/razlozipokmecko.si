package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


type Page struct {
    Title string
    Body []byte
}


const EDIT_PATH = "/edit/"
const VIEW_PATH = "/view/"
const SAVE_PATH = "/save/"


var pattern = filepath.Join("tmpl", "*.html")
var templates = template.Must(template.ParseGlob(pattern))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")


func main() {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    uri := os.Getenv("MONGODB_URI")

    client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))

    if err != nil {
        panic(err)
    }

    coll := client.Database("razlozipokmecko").Collection("explanations")

    cur, err := coll.Find(context.TODO(), bson.D{})

    if err == mongo.ErrNoDocuments {
        fmt.Printf("No documents found")
        return
    }

    if err != nil {
        panic(err)
    }

    defer cur.Close(context.TODO())

    var results []bson.M

    if err = cur.All(context.TODO(), &results); err != nil {
        log.Fatal(err)
    }

    for _, result := range results {
        fmt.Println(result)
    }

    http.HandleFunc(VIEW_PATH, makeHandler(viewHandler))
    http.HandleFunc(EDIT_PATH, makeHandler(editHandler))
    http.HandleFunc(SAVE_PATH, makeHandler(saveHandler))
    fmt.Println("Server running on port 8000")
    log.Fatal(http.ListenAndServe(":8000", nil))
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            return
        }
        fn(w, r, m[2])
    }
}


func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, EDIT_PATH + title, http.StatusFound)
        return
    }
    p.Title = strings.Join(strings.Split(p.Title, "-"), " ")
    renderTemplate(w, "view", p)
}


func editHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
        // if page doesn't exists create one
        p = &Page{Title: title}
    }
    p.Title = strings.Join(strings.Split(p.Title, "-"), " ")
    renderTemplate(w, "edit", p)
}


func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
    body := r.FormValue("body")
    p := &Page{Title: title, Body: []byte(body)}
    err := p.savePage()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, VIEW_PATH + title, http.StatusFound)
}


func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := templates.ExecuteTemplate(w, tmpl+".html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}


func loadPage(title string) (*Page, error) {
    filename := title + ".txt"
    body, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    return &Page{Title: title, Body: body}, nil
}

func (p *Page) savePage() error {
    filename := p.Title + ".txt"
    return os.WriteFile(filename, p.Body, 0600)
}
