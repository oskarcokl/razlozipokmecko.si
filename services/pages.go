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

type PageService struct {
	ms *db.MongoStore
}

func NewPageService(ms *db.MongoStore) *PageService {
	return &PageService{ms: ms}
}


func (ps *PageService) LoadPage(name string) (*m.Page, error) {
    var result m.Page
    filter := bson.D{{Key: "name", Value: name}}
    err := ps.ms.Coll.FindOne(context.TODO(), filter).Decode(&result)

    if err != nil {
        return nil, err
    }

    return &result, nil
}


func (ps *PageService) SavePage(p *m.Page) (*m.Page, error) {
    opts := options.Update().SetUpsert(true)

    oldName := p.Name
    name := strings.ToLower(strings.Join(strings.Split(p.Title, " "), "-"))

    if name != oldName {
        // I guess wi don't really need to check and just overwrite the value
        // always. But this generally allows us to change a title of an
        // explanation
        p.Name = name
    }

    update := bson.D{{Key: "$set", Value: p}}
    filter := bson.D{{Key: "name", Value: oldName}}
    result, err := ps.ms.Coll.UpdateOne(context.TODO(), filter, update, opts)

    if err != nil {
        log.Fatal(err)
        return nil, err
    }

    if result.MatchedCount != 0 {
        fmt.Println("Matched and replaced existing document")
        return p, nil
    }

    if result.UpsertedCount != 0 {
        fmt.Printf("Inserted a new document with ID %v\n", result.UpsertedID)
    }

    return p, nil
}


func (ps *PageService) LoadAllPages() ([]m.Page) {
    cur, err := ps.ms.Coll.Find(context.TODO(), bson.D{})

    if err != nil {
        log.Fatal(err)
    }

    var results []m.Page
    if err = cur.All(context.TODO(), &results); err != nil {
        log.Fatal(err)
    }

    return results
}