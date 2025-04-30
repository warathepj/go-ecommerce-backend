package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Sku represents a stock keeping unit in the database
type Sku struct {
	ProductID     int    `json:"productId" bson:"productId"`
	Sku           string `json:"sku" bson:"sku"`
	StockQuantity int    `json:"stockQuantity" bson:"stockQuantity"`
	CreatedAt     time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt" bson:"updatedAt"`
}

// handleSkus processes SKU-related requests
func handleSkus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
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