package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/oskarcokl/razlozipokmecko.si/db"
	markdown_parser "github.com/oskarcokl/razlozipokmecko.si/internal"
	m "github.com/oskarcokl/razlozipokmecko.si/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ExplanationService struct {
	ms *db.MongoStore
    Explanations []*m.Explanation
}


func NewPageService(ms *db.MongoStore) *ExplanationService {
    var explanations []*m.Explanation

	es := ExplanationService{
        ms: ms,
        Explanations: explanations,
    }

    es.LoadAllExplanations()

    return &es
}


func (es *ExplanationService) LoadExplanation(name string) (*m.Explanation, error) {
    for _, e := range(es.Explanations) {
        if e.Name == name {
            return e, nil
        }
    }

    return nil, fmt.Errorf("no explanation with name exists");
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


func (es *ExplanationService) LoadAllExplanations() (error) {
    dirPath := "./explanations"

    err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if info.IsDir() {
            return nil
        }

        data, err := os.ReadFile(path)
        if err != nil {
            return err
        }

        parts := strings.SplitN(string(data), "---", 3)
        if len(parts) < 3 {
            return fmt.Errorf("invalid markdown format")
        }

        metadata, err := markdown_parser.ParseMetadata(parts[1])
        if err != nil {
            return err
        }

        body := markdown_parser.ParseBody(parts[2])

        explanation := m.Explanation{
            Body: body,
            Name: strings.TrimSuffix(info.Name(), ".md"),
            Title: metadata.Title,
        }

        es.Explanations = append(es.Explanations, &explanation)

        return nil
    })

    if err != nil {
        return err
    }

    return nil
}