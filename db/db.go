package db

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


type MongoStore struct {
	Db *mongo.Client
	Coll *mongo.Collection
}

func NewMongoConnection(uri string) *MongoStore {
        client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))

        if err != nil {
            panic(err)
        }

        coll := client.Database("razlozipokmecko").Collection("explanations")

		ms := &MongoStore{
			Db: client,
			Coll: coll,
		}

		return ms
}

