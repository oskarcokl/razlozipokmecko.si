package services

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/oskarcokl/razlozipokmecko.si/db"
	m "github.com/oskarcokl/razlozipokmecko.si/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ExplanationService struct {
	ms *db.MongoStore
}

func NewPageService(ms *db.MongoStore) *ExplanationService {
	return &ExplanationService{ms: ms}
}


func (es *ExplanationService) LoadExplanation(name string) (*m.Explanation, error) {
    var result m.Explanation
    filter := bson.D{{Key: "name", Value: name}}
    err := es.ms.Coll.FindOne(context.TODO(), filter).Decode(&result)

    if err != nil {
        return nil, err
    }

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