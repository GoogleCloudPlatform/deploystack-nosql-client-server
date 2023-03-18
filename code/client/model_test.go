package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mockCollection struct {
	forceErr bool
	empty    bool
}

func (m mockCollection) InsertMany(context.Context, []interface{}, ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {

	result := mongo.InsertManyResult{}

	if m.forceErr {
		return nil, errForced
	}
	return &result, nil

}

func (m mockCollection) InsertOne(context.Context, interface{}, ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {

	result := mongo.InsertOneResult{}

	if m.forceErr {
		return nil, errForced
	}
	return &result, nil

}

func (m mockCollection) DeleteOne(context.Context, interface{}, ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	result := mongo.DeleteResult{}

	if m.forceErr {
		return nil, errForced
	}
	return &result, nil

}

func (m mockCollection) ReplaceOne(context.Context, interface{}, interface{}, ...*options.ReplaceOptions) (*mongo.UpdateResult, error) {
	result := mongo.UpdateResult{}

	if m.forceErr {
		return nil, errForced
	}
	return &result, nil

}

func (m mockCollection) Find(context.Context, interface{}, ...*options.FindOptions) (cur *mongo.Cursor, err error) {
	documents := []interface{}{}

	if !m.empty {
		documents = append(documents, trainer{Name: "Ash", Age: 20, City: "Pallet Town"})
		documents = append(documents, trainer{Name: "Misty", Age: 22, City: "Cerulean City"})
		documents = append(documents, trainer{Name: "Brock", Age: 35, City: "Pewter City"})
	}

	result, err := mongo.NewCursorFromDocuments(documents, nil, nil)

	if m.forceErr {
		return nil, errForced
	}
	return result, nil

}

func TestCollection(t *testing.T) {
	tests := map[string]struct {
		collection trainerCollection
		f          func(*trainerService) error
		err        error
	}{
		"load": {
			f: func(svc *trainerService) error {
				err := svc.load(context.Background(), []trainer{
					{Name: "Ash", Age: 20, City: "Pallet Town"},
					{Name: "Misty", Age: 22, City: "Cerulean City"},
					{Name: "Brock", Age: 35, City: "Pewter City"},
				})
				return err
			},
			collection: mockCollection{},
		},
		"loaderr": {
			f: func(svc *trainerService) error {
				err := svc.load(context.Background(), []trainer{
					{Name: "Ash", Age: 20, City: "Pallet Town"},
					{Name: "Misty", Age: 22, City: "Cerulean City"},
					{Name: "Brock", Age: 35, City: "Pewter City"},
				})
				return err
			},
			collection: mockCollection{forceErr: true},
			err:        errForced,
		},
		"loadempty": {
			f: func(svc *trainerService) error {
				err := svc.load(context.Background(), []trainer{
					{Name: "Ash", Age: 20, City: "Pallet Town"},
					{Name: "Misty", Age: 22, City: "Cerulean City"},
					{Name: "Brock", Age: 35, City: "Pewter City"},
				})
				return err
			},
			collection: mockCollection{empty: true},
		},

		"create": {
			f: func(svc *trainerService) error {
				t := trainer{Name: "Ash", Age: 20, City: "Pallet Town"}
				err := svc.create(context.Background(), t)
				return err

			},
			collection: mockCollection{},
		},
		"createerr": {
			f: func(svc *trainerService) error {
				t := trainer{Name: "Ash", Age: 20, City: "Pallet Town"}
				err := svc.create(context.Background(), t)
				return err

			},
			err:        errForced,
			collection: mockCollection{forceErr: true},
		},

		"delete": {
			f: func(svc *trainerService) error {
				t := trainer{Name: "Ash", Age: 20, City: "Pallet Town"}
				err := svc.delete(context.Background(), t)
				return err

			},
			collection: mockCollection{},
		},
		"deleteerr": {
			f: func(svc *trainerService) error {
				t := trainer{Name: "Ash", Age: 20, City: "Pallet Town"}
				err := svc.delete(context.Background(), t)
				return err

			},
			err:        errForced,
			collection: mockCollection{forceErr: true},
		},

		"update": {
			f: func(svc *trainerService) error {
				t := trainer{Name: "Ash", Age: 20, City: "Pallet Town"}
				t2 := trainer{Name: "Ash2", Age: 20, City: "Pallet Town"}
				err := svc.update(context.Background(), t, t2)
				return err

			},
			collection: mockCollection{},
		},
		"updateerr": {
			f: func(svc *trainerService) error {
				t := trainer{Name: "Ash", Age: 20, City: "Pallet Town"}
				t2 := trainer{Name: "Ash2", Age: 20, City: "Pallet Town"}
				err := svc.update(context.Background(), t, t2)
				return err

			},
			err:        errForced,
			collection: mockCollection{forceErr: true},
		},
		"init": {
			f: func(svc *trainerService) error {
				err := initData(context.Background(), svc)
				return err

			},
			collection: mockCollection{},
		},
		"initerr": {
			f: func(svc *trainerService) error {
				err := initData(context.Background(), svc)
				return err

			},
			err:        errForced,
			collection: mockCollection{forceErr: true},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			svc := &trainerService{
				collection: tc.collection,
			}

			err := tc.f(svc)
			if tc.err == nil {
				require.Nil(t, err)
			} else {
				assert.ErrorIs(t, err, tc.err)
			}

		})
	}
}

func TestNewTrainerService(t *testing.T) {
	dl := 1 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), dl)
	defer cancel()

	_, err := newTrainerService(ctx, "", "")
	want := "error initializing data"

	assert.ErrorContains(t, err, want)

}
