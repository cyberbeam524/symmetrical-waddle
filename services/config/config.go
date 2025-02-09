package config

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/stripe/stripe-go/v72"
)

// MongoDB instance
var MongoDB *mongo.Database

// Google and Stripe configuration variables
var GoogleClientID string
var StripeSecretKey string
var StripeSuccessURL string
var StripeCancelURL string
var BunnyNetUploadURL string

// Setup initializes the configuration for MongoDB and Stripe
func Setup() {
	// Initialize MongoDB
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGODB_URI"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	MongoDB = client.Database(os.Getenv("MONGODB_DATABASE"))

	// Initialize Stripe
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Load Google OAuth client ID and Stripe settings from environment variables
	GoogleClientID = os.Getenv("GOOGLE_CLIENT_ID")
	StripeSecretKey = os.Getenv("STRIPE_SECRET_KEY")
	StripeSuccessURL = os.Getenv("STRIPE_SUCCESS_URL")
	StripeCancelURL = os.Getenv("STRIPE_CANCEL_URL")
	BunnyNetUploadURL = os.Getenv("BUNNYNET_URL")
}
