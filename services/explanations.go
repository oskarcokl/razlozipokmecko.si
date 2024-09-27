package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/oskarcokl/razlozipokmecko.si/db"
	m "github.com/oskarcokl/razlozipokmecko.si/models"
	"github.com/russross/blackfriday/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v2"
)

type ExplanationService struct {
	ms *db.MongoStore
}

type ExplanationMetaData struct {
    Title string `yaml:"title"`
    DateCreated string `yaml:"date_created"`
}

func NewPageService(ms *db.MongoStore) *ExplanationService {
	return &ExplanationService{ms: ms}
}


func (es *ExplanationService) LoadExplanation(name string) (*m.Explanation, error) {
    var result m.Explanation

    data, err := os.ReadFile("./explanations/" + name + ".md")
    if err != nil {
        return nil, err
    }

    parts := strings.SplitN(string(data), "---", 3)
    if len(parts) < 3 {
        return nil, fmt.Errorf("invalid markdown format")
    }

    var metadata ExplanationMetaData
    err = yaml.Unmarshal([]byte(parts[1]), &metadata)
    if err != nil {
        return nil, err
    }

    body := blackfriday.Run([]byte(parts[2]))

    result.Body = body
    result.Name = name
    result.Title = metadata.Title

    return &result, nil
}


func (es *ExplanationService) SaveExplanation(e *m.Explanation) (*m.Explanation, error) {
    opts := options.Update().SetUpsert(true)

    oldName := e.Name
    name := strings.ToLower(strings.Join(strings.Split(e.Title, " "), "-"))

    if name != oldName {
        // I guess wi don't really need to check and just overwrite the value
        // always. But this generally allows us to change a title of an
        // explanation
        e.Name = name
    }

    update := bson.D{{Key: "$set", Value: e}}
    filter := bson.D{{Key: "name", Value: oldName}}
    result, err := es.ms.Coll.UpdateOne(context.TODO(), filter, update, opts)

    if err != nil {
        log.Fatal(err)
        return nil, err
    }

    if result.MatchedCount != 0 {
        fmt.Println("Matched and replaced existing document")
        return e, nil
    }

    if result.UpsertedCount != 0 {
        fmt.Printf("Inserted a new document with ID %v\n", result.UpsertedID)
    }

    return e, nil
}


func (es *ExplanationService) LoadAllExplanations() ([]m.Explanation) {
    cur, err := es.ms.Coll.Find(context.TODO(), bson.D{})

    if err != nil {
        log.Fatal(err)
    }

    var results []m.Explanation
    if err = cur.All(context.TODO(), &results); err != nil {
        log.Fatal(err)
    }

    return results
}