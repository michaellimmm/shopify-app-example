package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/zeals-co-ltd/shopify-app-example/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	authCollection = "auth"
)

type AuthRepository interface {
	FindAll(ctx context.Context) ([]model.ShopifyAuth, error)
	FindByShop(ctx context.Context, shop string) (model.ShopifyAuth, error)
	Save(ctx context.Context, data model.ShopifyAuth) (model.ShopifyAuth, error)
}

type authRepository struct {
	collection *mongo.Collection
}

func NewAuthRepository(db *mongo.Database) (AuthRepository, error) {
	collection := db.Collection(authCollection)
	if collection == nil {
		return nil, fmt.Errorf("failed to get collection %s", authCollection)
	}

	return &authRepository{
		collection: collection,
	}, nil
}

func (r *authRepository) FindAll(ctx context.Context) ([]model.ShopifyAuth, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return []model.ShopifyAuth{}, err
	}

	var results []model.ShopifyAuth
	err = cursor.All(ctx, &results)
	if err != nil {
		return []model.ShopifyAuth{}, err

	}

	return results, nil
}

func (r *authRepository) FindByShop(ctx context.Context, shop string) (model.ShopifyAuth, error) {
	filter := bson.M{}
	filter["shop"] = shop

	var result model.ShopifyAuth
	err := r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.ShopifyAuth{}, nil
		}
		return model.ShopifyAuth{}, err
	}

	return result, nil
}

func (r *authRepository) Save(ctx context.Context, data model.ShopifyAuth) (model.ShopifyAuth, error) {
	data.SetID()
	data.UpdateDate()

	_, err := r.collection.InsertOne(ctx, &data)
	if err != nil {
		return model.ShopifyAuth{}, err
	}

	return data, nil
}
