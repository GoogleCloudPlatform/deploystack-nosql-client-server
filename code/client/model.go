package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type trainer struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	City string `json:"city"`
}

type trainerCRUDer interface {
	load(context.Context, []trainer) error
	list(context.Context) ([]*trainer, error)
	create(context.Context, trainer) error
	delete(context.Context, trainer) error
	update(context.Context, trainer, trainer) error
}

type trainerCollection interface {
	InsertMany(context.Context, []interface{}, ...*options.InsertManyOptions) (*mongo.InsertManyResult, error)
	InsertOne(context.Context, interface{}, ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	DeleteOne(context.Context, interface{}, ...*options.DeleteOptions) (*mongo.DeleteResult, error)
	ReplaceOne(context.Context, interface{}, interface{}, ...*options.ReplaceOptions) (*mongo.UpdateResult, error)
	Find(context.Context, interface{}, ...*options.FindOptions) (cur *mongo.Cursor, err error)
}

type trainerService struct {
	collection trainerCollection
}

// newTrainerService spins up a new TrainerManager for interacting with MongoDB.
func newTrainerService(ctx context.Context, host, port string) (*trainerService, error) {
	uri := fmt.Sprintf("mongodb://%s:%s", host, port)
	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("error connecting to mongo: %w", err)
	}

	collection := client.Database("test").Collection("trainers")

	svc := &trainerService{
		collection: collection,
	}

	if err := initData(ctx, svc); err != nil {
		return nil, fmt.Errorf("error initializing data: %w", err)
	}

	return svc, nil
}

func initData(ctx context.Context, svc trainerCRUDer) error {
	trainers := []trainer{
		{Name: "Ash", Age: 20, City: "Pallet Town"},
		{Name: "Misty", Age: 22, City: "Cerulean City"},
		{Name: "Brock", Age: 35, City: "Pewter City"},
	}

	return svc.load(ctx, trainers)
}

// load pushes a collection of trainers into a mongoDB instance
func (svc *trainerService) load(ctx context.Context, trainers []trainer) error {
	t := make([]interface{}, len(trainers))
	for i, tdata := range trainers {
		t[i] = tdata
	}

	list, err := svc.list(ctx)
	if err != nil {
		return fmt.Errorf("error checking before loading to mongo: %w", err)
	}

	if len(list) > 0 {
		return nil
	}

	if _, err := svc.collection.InsertMany(ctx, t); err != nil {
		return fmt.Errorf("error inserting records to mongo: %w", err)
	}

	return nil
}

func (svc *trainerService) create(ctx context.Context, trainer trainer) error {

	if _, err := svc.collection.InsertOne(ctx, trainer); err != nil {
		return fmt.Errorf("error inserting record to mongo: %w", err)
	}

	return nil
}

func (svc *trainerService) delete(ctx context.Context, trainer trainer) error {
	if _, err := svc.collection.DeleteOne(ctx, trainer); err != nil {
		return fmt.Errorf("error inserting record to mongo: %w", err)
	}
	return nil
}

func (svc *trainerService) update(ctx context.Context, original, replacement trainer) error {
	if _, err := svc.collection.ReplaceOne(ctx, original, replacement); err != nil {
		return fmt.Errorf("error replacing record in mongo: %w", err)
	}
	return nil
}

// list retrieves the total collection of trainers from a mongoDB instance
func (svc *trainerService) list(ctx context.Context) ([]*trainer, error) {
	var results []*trainer

	cur, err := svc.collection.Find(ctx, bson.D{{}}, options.Find())
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var elem trainer
		if err := cur.Decode(&elem); err != nil {
			return nil, err
		}
		results = append(results, &elem)
	}

	return results, cur.Err()
}
