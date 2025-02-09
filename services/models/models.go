package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)


// User model
type User struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty"`
	Email                string             `bson:"email"`
	GoogleID             string             `bson:"google_id,omitempty"`
	StripeID             string             `bson:"stripe_id,omitempty"`
	Subscription         string             `bson:"subscription"`
	Uploads              int                `bson:"uploads"`
	IsPremium            bool               `bson:"is_premium"`
	Credits              int                `bson:"credits"`
	LastProcessedSession string             `bson:"last_processed_session,omitempty"` // New field
}