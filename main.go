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
const LIST_VIEW_PATH = "/list-view/"


var pattern = filepath.Join("tmpl", "*.html")
var templates = template.Must(template.ParseGlob(pattern))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")


func main() {
    http.HandleFunc(VIEW_PATH, makeHandler(viewHandler))
    http.HandleFunc(EDIT_PATH, makeHandler(editHandler))
    http.HandleFunc(SAVE_PATH, makeHandler(saveHandler))
    http.HandleFunc(LIST_VIEW_PATH, listViewHandler)
    fmt.Println("Server running on port 8000")
    log.Fatal(http.ListenAndServe(":8000", nil))
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string, *mongo.Collection)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            return
        }

        if err := godotenv.Load(); err != nil {
            log.Println("No .env file found")
        }

        uri := os.Getenv("MONGODB_URI")

        client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))

        if err != nil {
            panic(err)
        }

        coll := client.Database("razlozipokmecko").Collection("explanations")

        fn(w, r, m[2], coll)
    }
}


func listViewHandler(w http.ResponseWriter, r *http.Request) {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    uri := os.Getenv("MONGODB_URI")

    client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))

    if err != nil {
        panic(err)
    }

    coll := client.Database("razlozipokmecko").Collection("explanations")

    pages := loadAllPages(coll)

    err = templates.ExecuteTemplate(w, "list-view.html", pages)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}


func viewHandler(w http.ResponseWriter, r *http.Request, title string, coll *mongo.Collection) {
    p, err := loadPage(strings.Join(strings.Split(title, "-"), " "), coll)
    if err != nil {
        http.Redirect(w, r, EDIT_PATH + title, http.StatusFound)
        return
    }
    renderTemplate(w, "view", p)
}


func editHandler(w http.ResponseWriter, r *http.Request, title string, coll *mongo.Collection) {
    p, err := loadPage(title, coll)
    if err != nil {
        // if page doesn't exists create one
        p = &Page{Title: title}
    }
    p.Title = strings.Join(strings.Split(p.Title, "-"), " ")
    renderTemplate(w, "edit", p)
}


func saveHandler(w http.ResponseWriter, r *http.Request, title string, coll *mongo.Collection) {
    body := r.FormValue("body")
    p := &Page{Title: title, Body: []byte(body)}
    err := p.savePage(coll)
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


func loadPage(title string, coll *mongo.Collection) (*Page, error) {
    var result Page
    err := coll.FindOne(context.TODO(), bson.D{{"title", title}}).Decode(&result)

    if err != nil {
        return nil, err
    }

    return &result, nil
}


func loadAllPages(coll *mongo.Collection) ([]Page) {
    cur, err := coll.Find(context.TODO(), bson.D{})

    if err != nil {
        log.Fatal(err)
    }

    var results []Page
    if err = cur.All(context.TODO(), &results); err != nil {
        log.Fatal(err)
    }

    return results
}


func (p *Page) savePage(coll *mongo.Collection) error {
    opts := options.Update().SetUpsert(true)
    update := bson.D{{"$set", p}}
    result, err := coll.UpdateOne(context.TODO(), bson.D{{"title", p.Title}}, update, opts)

    if err != nil {
        log.Fatal(err)
        return err
    }

    if result.MatchedCount != 0 {
        fmt.Println("Matched and replaced existing document")
        return nil
    }

    if result.UpsertedCount != 0 {
        fmt.Printf("Inserted a new document with ID %v\n", result.UpsertedID)
    }

    return nil
}
