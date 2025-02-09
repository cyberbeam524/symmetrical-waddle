package controllers

import (
	"encoding/json"
	"context"
	"net/http"
	"services/scraper/config"
	"services/scraper/models"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/webhook"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"github.com/stripe/stripe-go/v72/sub"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo" // Reintroduced for ErrNoDocuments

	"strconv" // Added for string conversion
	"fmt"     // Added for formatted printing/logging
	"io/ioutil"
	"log"
)


func CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
    err := r.ParseForm()
    if err != nil {
        http.Error(w, "Failed to parse form", http.StatusBadRequest)
        return
    }
    email := r.FormValue("email")
    credits := r.FormValue("credits") // Get the number of credits from the request

    // Validate credits
    creditAmount, err := strconv.Atoi(credits)
    if err != nil || creditAmount <= 0 {
        http.Error(w, "Invalid credit amount", http.StatusBadRequest)
        return
    }

    stripe.Key = config.StripeSecretKey

    // // Create a checkout session for credits
    // params := &stripe.CheckoutSessionParams{
    //     PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
    //     LineItems: []*stripe.CheckoutSessionLineItemParams{
    //         {
    //             PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
    //                 Currency: stripe.String(string(stripe.CurrencyUSD)),
    //                 ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
    //                     Name: stripe.String(fmt.Sprintf("%d Credits", creditAmount)),
    //                 },
    //                 UnitAmount: stripe.Int64(int64(creditAmount * 100)), // $1 per credit
    //             },
    //             Quantity: stripe.Int64(1),
    //         },
    //     },
    //     Mode:          stripe.String(string(stripe.CheckoutSessionModePayment)),
    //     // SuccessURL:    stripe.String(config.StripeSuccessURL),
    //     // CancelURL:     stripe.String(config.StripeCancelURL),
	// 	SuccessURL: stripe.String(fmt.Sprintf("http://localhost:3000/credit?email=%s&credits=%d", email, creditAmount)),
	// 	CancelURL:  stripe.String(config.StripeCancelURL),
    //     CustomerEmail: stripe.String(email),
    // }

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(string(stripe.CurrencyUSD)),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(fmt.Sprintf("%d Credits", creditAmount)),
					},
					UnitAmount: stripe.Int64(int64(creditAmount * 100)), // $1 per credit
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:          stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:    stripe.String(config.StripeSuccessURL),
		CancelURL:     stripe.String(config.StripeCancelURL),
		CustomerEmail: stripe.String(email),
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: map[string]string{
				"credits": fmt.Sprintf("%d", creditAmount),
			},
		},
	}

    s, err := session.New(params)
    if err != nil {
        http.Error(w, "Failed to create checkout session: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"sessionId": s.ID})
}


func CreateCheckoutSession2(w http.ResponseWriter, r *http.Request) {
	// Parse form to extract email
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	email := r.FormValue("email")

	// Set the Stripe secret key
	stripe.Key = config.StripeSecretKey

	// Create a checkout session with recurring price (subscription mode)
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(string(stripe.CurrencyUSD)),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Premium Subscription"),
					},
					UnitAmount: stripe.Int64(1000), // $10 per month
					Recurring: &stripe.CheckoutSessionLineItemPriceDataRecurringParams{
						Interval: stripe.String("month"),
					},
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:           stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL:     stripe.String(config.StripeSuccessURL),
		CancelURL:      stripe.String(config.StripeCancelURL),
		CustomerEmail:  stripe.String(email),
	}

	// Create a new Stripe checkout session
	s, err := session.New(params)
	if err != nil {
		http.Error(w, "Failed to create checkout session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the session ID to the client
	response := map[string]string{"sessionId": s.ID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


// but what if this service is down and the stripe server unable to access this url to get to this method? 
// then the correct credit wouoldnt be added? - doesnt matter once service is back up can check if all the checkout session ids for that user from stripe 
// Retrieve the Checkout Session from the API with the line_items property expanded.


// https://docs.stripe.com/error-handling
func HandlePaymentSuccess(w http.ResponseWriter, r *http.Request) {
    // Extract query parameters
    email := r.URL.Query().Get("email")
    credits := r.URL.Query().Get("credits")
    
    if email == "" || credits == "" {
        http.Error(w, "Missing email or credits", http.StatusBadRequest)
        return
    }

    // Validate and convert credits to an integer
    creditAmount, err := strconv.Atoi(credits)
    if err != nil || creditAmount <= 0 {
        http.Error(w, "Invalid credit amount", http.StatusBadRequest)
        return
    }

    // Find the user in the database
    var user models.User
    err = config.MongoDB.Collection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            http.Error(w, "User not found", http.StatusNotFound)
        } else {
            http.Error(w, "Failed to fetch user", http.StatusInternalServerError)
        }
        return
    }

    // Increment the user's credits
    update := bson.M{"$inc": bson.M{"credits": creditAmount}}
    _, err = config.MongoDB.Collection("users").UpdateOne(context.Background(), bson.M{"email": email}, update)
    if err != nil {
        http.Error(w, "Failed to update user credits", http.StatusInternalServerError)
        return
    }

    // Send success response
    response := map[string]string{
        "status":  "success",
        "message": "Payment successful, credits added",
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}


func HandlePaymentSuccess3(w http.ResponseWriter, r *http.Request) {
    email := r.URL.Query().Get("email")
    credits := r.URL.Query().Get("credits") // Pass credits as a query parameter
    creditAmount, err := strconv.Atoi(credits)
    if err != nil || creditAmount <= 0 {
        http.Error(w, "Invalid credit amount", http.StatusBadRequest)
        return
    }

    var user models.User
    err = config.MongoDB.Collection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
    if err != nil {
        http.Error(w, "User not found", http.StatusInternalServerError)
        return
    }

    // Update user credits
    update := bson.M{
        "$inc": bson.M{"credits": creditAmount},
    }
    _, err = config.MongoDB.Collection("users").UpdateOne(context.Background(), bson.M{"email": email}, update)
    if err != nil {
        http.Error(w, "Failed to update user credits", http.StatusInternalServerError)
        return
    }

    response := map[string]string{"status": "payment successful, credits added"}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}



func HandlePaymentSuccess2(w http.ResponseWriter, r *http.Request) {
	// Extract email and stripeSubscriptionID from query parameters
	email := r.URL.Query().Get("email")
	stripeSubscriptionID := r.URL.Query().Get("stripeSubscriptionID")

	var user models.User
	err := config.MongoDB.Collection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	// Update user to premium status after successful payment
	update := bson.M{
		"$set": bson.M{
			"subscription": "premium",
			"is_premium":   true,
			"stripe_id":    stripeSubscriptionID,
		},
	}
	_, err = config.MongoDB.Collection("users").UpdateOne(context.Background(), bson.M{"email": email}, update)
	if err != nil {
		http.Error(w, "Could not update user subscription", http.StatusInternalServerError)
		return
	}

	response := map[string]string{"status": "payment successful, user upgraded to premium"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func DeductCredits(userEmail string, amount int) error {
    result := config.MongoDB.Collection("users").FindOneAndUpdate(
        context.Background(),
        bson.M{"email": userEmail, "credits": bson.M{"$gte": amount}}, // Ensure sufficient credits
        bson.M{"$inc": bson.M{"credits": -amount}},
    )
    if result.Err() != nil {
        return fmt.Errorf("insufficient credits or user not found")
    }
    return nil
}


func CancelSubscription(w http.ResponseWriter, r *http.Request) {
	// Parse form to extract email
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	email := r.FormValue("email")

	// Fetch the user's data from MongoDB
	var user models.User
	err = config.MongoDB.Collection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if the user has a Stripe subscription ID
	if user.StripeID == "" {
		http.Error(w, "No active subscription", http.StatusBadRequest)
		return
	}

	// Set the Stripe secret key
	stripe.Key = config.StripeSecretKey

	// Cancel the subscription using the Stripe API
	params := &stripe.SubscriptionCancelParams{}
	_, err = sub.Cancel(user.StripeID, params)
	if err != nil {
		http.Error(w, "Failed to cancel subscription", http.StatusInternalServerError)
		return
	}

	// Update the user's subscription status in the database
	update := bson.M{
		"$set": bson.M{
			"subscription": "none",
			"is_premium":   false,
			"stripe_id":    "",
		},
	}
	_, err = config.MongoDB.Collection("users").UpdateOne(context.Background(), bson.M{"email": email}, update)
	if err != nil {
		http.Error(w, "Failed to update user status", http.StatusInternalServerError)
		return
	}

	response := map[string]string{"status": "subscription canceled"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetUserCredits(w http.ResponseWriter, r *http.Request) {
    email := r.URL.Query().Get("email")
    if email == "" {
        http.Error(w, "Email is required", http.StatusBadRequest)
        return
    }

    var user models.User
    err := config.MongoDB.Collection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            http.Error(w, "User not found", http.StatusNotFound)
        } else {
            http.Error(w, "Failed to fetch user", http.StatusInternalServerError)
        }
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{"credits": user.Credits})
}

const webhookSecret = "whsec_08c705339f26433299b5f5fd9a3a0f7b01781906d11fbde718b802f28e4dd5e3" // Replace with your webhook secret
// const webhookSecretLive = "whsec_5GOx8l1QFn75N83QpNcoAxw4qV6Jhmoo"


func StripeWebhookHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("StripeWebhookHandler")
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusBadRequest)
		return
	}

	// Verify the webhook signature
	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), webhookSecret)
	if err != nil {
		http.Error(w, "Webhook signature verification failed", http.StatusBadRequest)
		return
	}

	// Handle the event
	if event.Type == "checkout.session.completed" {
		log.Printf("checkout.session.completed")
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			log.Printf("Failed to parse event data")
			http.Error(w, "Failed to parse event data", http.StatusBadRequest)
			return
		}

		email := session.CustomerEmail

		// Retrieve the PaymentIntent metadata
		paymentIntentID := session.PaymentIntent.ID
		paymentIntent, err := paymentintent.Get(paymentIntentID, nil)
		if err != nil {
			log.Printf("Failed to retrieve PaymentIntent")
			http.Error(w, "Failed to retrieve PaymentIntent", http.StatusInternalServerError)
			return
		}

		creditsStr := paymentIntent.Metadata["credits"]
		credits, err := strconv.Atoi(creditsStr)
		if err != nil || credits <= 0 {
			log.Printf("Invalid credit amount: %s", creditsStr)
			http.Error(w, "Invalid credit amount", http.StatusBadRequest)
			return
		}

		// Check if payment was already processed
		var user models.User
		err = config.MongoDB.Collection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
		if err != nil {
			log.Printf("User not found: %s", email)
			http.Error(w, "User not found", http.StatusInternalServerError)
			return
		}

		// Prevent double processing
		if user.LastProcessedSession == session.ID {
			w.WriteHeader(http.StatusOK)
			log.Printf("Payment already processed: %s", email)
			w.Write([]byte("Payment already processed"))
			return
		}

		// Increment credits
		update := bson.M{
			"$inc": bson.M{"credits": credits},
			"$set": bson.M{"last_processed_session": session.ID},
		}
		_, err = config.MongoDB.Collection("users").UpdateOne(context.Background(), bson.M{"email": email}, update)
		if err != nil {
			log.Printf("Failed to update user credits: %s", email)
			http.Error(w, "Failed to update user credits", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		log.Printf("Credits updated successfully: %s", email)
		w.Write([]byte("Credits updated successfully"))
		return
	}

	// Return 200 for unsupported events
	w.WriteHeader(http.StatusOK)
}
