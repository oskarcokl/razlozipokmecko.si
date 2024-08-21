package services

import (
	"context"
	"fmt"
	"log"

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


func (ps *PageService) LoadPage(title string) (*m.Page, error) {
    var result m.Page
    err := ps.ms.Coll.FindOne(context.TODO(), bson.D{{"title", title}}).Decode(&result)

    if err != nil {
        return nil, err
    }

    return &result, nil
}


func (ps *PageService) SavePage(p *m.Page) (error) {
    opts := options.Update().SetUpsert(true)
    update := bson.D{{"$set", p}}
    result, err := ps.ms.Coll.UpdateOne(context.TODO(), bson.D{{"title", p.Title}}, update, opts)

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