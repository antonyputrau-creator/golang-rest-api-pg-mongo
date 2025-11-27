package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Product struct {
    ID    primitive.ObjectID `bson:"_id,omitempty" json:"id"` // ObjectID [cite: 25]
    Name  string             `bson:"name" json:"name"`
    Price float64            `bson:"price" json:"price"` // float64 [cite: 26]
}