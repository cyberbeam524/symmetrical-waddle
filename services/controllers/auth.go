package controllers

import (
	"encoding/json"
	"context"
	"fmt"
	"net/http"
	"time"
	"services/scraper/config"
	"services/scraper/models"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/idtoken"
)

// JWT secret key
var jwtSecret = []byte("your-secret-key")

// Function to generate a JWT token
func GenerateJWT(userID string) (string, error) {
	// Define token claims
	claims := jwt.MapClaims{
		"user_id": userID,           // Store the user ID in the token
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Set token expiration (24 hours)
	}

	// Create a new token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}


var client = &http.Client{}
func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("id_token") // Extract token from form data

	// Verify Google ID token
	payload, err := idtoken.Validate(context.Background(), token, config.GoogleClientID)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	email := payload.Claims["email"].(string)
	var user models.User

	// Check if user exists in MongoDB
	err = config.MongoDB.Collection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		// If user does not exist, sign them up
		newUser := models.User{
			Email:       email,
			GoogleID:    payload.Subject,
			Subscription: "freemium",
			Uploads:      0,
			IsPremium:    false,
		}
		_, err = config.MongoDB.Collection("users").InsertOne(context.Background(), newUser)
		if err != nil {
			http.Error(w, "Could not create user", http.StatusInternalServerError)
			return
		}
		user = newUser
	}

	// Generate JWT token for the user
	jwtToken, err := GenerateJWT(user.ID.Hex())
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	// Return the JWT token in the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "logged in",
		"email":  user.Email,
		"token":  jwtToken,
	})
}



func EmailLogin(w http.ResponseWriter, r *http.Request) {
	// Parse the form to get the email
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	email := r.FormValue("email")
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}
	fmt.Println("Email received:", email)

	var user models.User

	// Check if the user exists in the MongoDB collection
	err = config.MongoDB.Collection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// If the user doesn't exist, create a new user
			fmt.Println("User not found. Creating new user...")
			newUser := models.User{
				Email:       email,
				Subscription: "freemium",
				Uploads:      0,
				IsPremium:    false,
			}
			_, insertErr := config.MongoDB.Collection("users").InsertOne(context.Background(), newUser)
			if insertErr != nil {
				http.Error(w, fmt.Sprintf("Could not create user: %v", insertErr), http.StatusInternalServerError)
				return
			}
			// Update the user variable to hold the newly created user data
			user = newUser
		} else {
			// Log the specific error and respond with a server error
			http.Error(w, fmt.Sprintf("Failed to fetch user: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Generate a JWT token for the user
	jwtToken, err := GenerateJWT(user.ID.Hex())
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not generate token: %v", err), http.StatusInternalServerError)
		return
	}

	// Create the response as a JSON object
	response := map[string]interface{}{
		"status": "logged in",
		"email":  user.Email,
		"token":  jwtToken, // Include the JWT token in the response
	}

	// Set the response headers and return the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

