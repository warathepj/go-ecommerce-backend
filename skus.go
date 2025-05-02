package main

import (
	"context"
	"encoding/json"
	"log" // Import the log package
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson" // Import bson package
	// "go.mongodb.org/mongo-driver/bson/primitive" // Commented out as it's not being used
	// "go.mongodb.org/mongo-driver/mongo" // Add this import for the mongo.Database type
	"go.mongodb.org/mongo-driver/mongo/options" // Added the missing import here
)

// Remove the db variable declaration since it's already declared in db.go
// var db *mongo.Database

// Sku represents a stock keeping unit in the database
type Sku struct {
	// Assuming ProductID in Sku corresponds to _id in Product, which might be an ObjectID or an int.
	// Adjust the type here and in the query logic if Product ID is not an int.
	// If Product._id is an ObjectID, use primitive.ObjectID here and handle potential conversion errors.
	// If Product._id is stored differently (e.g., custom string), adjust accordingly.
	// For now, keeping it as int based on the original struct.
	ProductID     int       `json:"productId" bson:"productId"`
	Sku           string    `json:"sku" bson:"sku"`
	StockQuantity int       `json:"stockQuantity" bson:"stockQuantity"`
	CreatedAt     time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt" bson:"updatedAt"`
}

// Define a simple Product struct just for retrieving the ID
// Adjust the type of ID based on how it's stored in MongoDB (e.g., primitive.ObjectID, int, string)
type ProductID struct {
	ID int `bson:"id"` // Changed back to "id" to match your actual data structure
}

// handleSkus processes SKU-related requests
func handleSkus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		log.Println("Starting SKU lookup for Wireless Mouse products")
		// Find products named "Wireless Mouse"
		productsCollection := db.Collection("products")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Define the filter to find products by name
		productFilter := bson.M{"name": "Wireless Mouse"}
		// Specify projection to only retrieve the ID field
		opts := options.Find().SetProjection(bson.M{"id": 1})

		log.Println("Executing query to find products with name 'Wireless Mouse'")
		cursor, err := productsCollection.Find(ctx, productFilter, opts)
		if err != nil {
			log.Printf("Error finding products: %v", err)
			http.Error(w, "Failed to query products", http.StatusInternalServerError)
			return
		}
		defer cursor.Close(ctx)

		var productIDs []int
		for cursor.Next(ctx) {
			var prod ProductID
			if err := cursor.Decode(&prod); err != nil {
				log.Printf("Error decoding product ID: %v", err)
				continue
			}
			productIDs = append(productIDs, prod.ID)
			log.Printf("Found product with ID: %d", prod.ID)
		}
		if err := cursor.Err(); err != nil {
			log.Printf("Error iterating product cursor: %v", err)
			http.Error(w, "Failed to process products", http.StatusInternalServerError)
			return
		}

		if len(productIDs) == 0 {
			log.Println("No products found with name 'Wireless Mouse'")
			json.NewEncoder(w).Encode([]Sku{}) // Return empty array
			return
		}

		log.Printf("Found %d product IDs for 'Wireless Mouse': %v", len(productIDs), productIDs)

		// Find SKUs matching the product IDs
		skusCollection := db.Collection("skus")
		skuFilter := bson.M{"productId": bson.M{"$in": productIDs}}

		log.Printf("Searching for SKUs with productId in: %v", productIDs)
		skuCursor, err := skusCollection.Find(ctx, skuFilter)
		if err != nil {
			log.Printf("Error finding skus: %v", err)
			http.Error(w, "Failed to query SKUs", http.StatusInternalServerError)
			return
		}
		defer skuCursor.Close(ctx)

		var skus []Sku
		if err = skuCursor.All(ctx, &skus); err != nil {
			log.Printf("Error decoding skus: %v", err)
			http.Error(w, "Failed to process SKUs", http.StatusInternalServerError)
			return
		}

		// Log all found SKUs with more detailed information
		log.Println("=== SKUs for Wireless Mouse ===")
		if len(skus) == 0 {
			log.Println("No SKUs found for Wireless Mouse products")
		} else {
			log.Printf("Found %d SKUs for Wireless Mouse products", len(skus))
			for i, sku := range skus {
				log.Printf("SKU #%d: ID=%d, SKU=%s, Stock=%d, Created=%v, Updated=%v",
					i+1, sku.ProductID, sku.Sku, sku.StockQuantity, sku.CreatedAt, sku.UpdatedAt)
			}
		}
		log.Println("===============================")

		// Return the found SKUs
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(skus)

	case "POST":
		var sku Sku
		if err := json.NewDecoder(r.Body).Decode(&sku); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Add timestamps
		sku.CreatedAt = time.Now()
		sku.UpdatedAt = time.Now()

		collection := db.Collection("skus")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := collection.InsertOne(ctx, sku)
		if err != nil {
			http.Error(w, "Failed to add SKU", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "SKU added successfully",
			"id":      result.InsertedID,
		})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
