package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ShopifyAuth struct {
	ID          primitive.ObjectID `bson:"_id"`
	Shop        string             `bson:"shop"`
	AccessToken string             `bson:"access_token"`
	CreatedAt   *time.Time         `bson:"created_at,omitempty"`
	UpdatedAt   *time.Time         `bson:"updated_at,omitempty"`
	DeletedAt   *time.Time         `bson:"deleted_at,omitempty"`
}

func (s ShopifyAuth) IsEmpty() bool {
	return s.ID.IsZero() &&
		s.Shop == "" &&
		s.AccessToken == "" &&
		s.CreatedAt == nil &&
		s.UpdatedAt == nil &&
		s.DeletedAt == nil
}

func (s *ShopifyAuth) SetID() {
	if s.ID.IsZero() {
		s.ID = primitive.NewObjectID()
	}
}

func (s *ShopifyAuth) UpdateDate() {
	now := time.Now()
	if s.CreatedAt == nil {
		s.CreatedAt = &now
	}

	s.UpdatedAt = &now
}
