package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/mongo"
)

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	// Initialize MongoDB connection
	if err := initMongoDB(); err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	// Log product with ID=2 when the server starts
	logProductWithID2()
	
	// Log all products when the server starts
	logAllProducts()

	// Register routes with CORS middleware
	http.HandleFunc("/", enableCORS(helloHandler))
	http.HandleFunc("/api/products", enableCORS(getProducts))
	http.HandleFunc("/api/orders", enableCORS(createOrder))
	http.HandleFunc("/api/skus", enableCORS(handleSkus)) // Add this line

	// Start server
	fmt.Println("Server starting on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// logProductWithID2 fetches and logs the product with ID=2 from the database
func logProductWithID2() {
	collection := client.Database("ecommerce").Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a filter for product with id=2
	filter := bson.M{"id": 2}
	
	// Find the product
	var product map[string]interface{}
	err := collection.FindOne(ctx, filter).Decode(&product)
	if err != nil {
		log.Printf("Error finding product with ID=2: %v", err)
		return
	}

	// Convert the product to JSON for pretty printing
	productJSON, err := json.MarshalIndent(product, "", "  ")
	if err != nil {
		log.Printf("Error marshaling product data: %v", err)
		return
	}

	// Log the product data
	log.Printf("=== Product with ID=2 ===\n%s\n=======================", string(productJSON))
}

// getProducts handles requests for product data
func getProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		products, err := fetchProductsFromDB()
		if err != nil {
			http.Error(w, "Failed to fetch products", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(products)
	case "POST":
		var product map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		collection := client.Database("ecommerce").Collection("products")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		result, err := collection.InsertOne(ctx, product)
		if err != nil {
			http.Error(w, "Failed to add product", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Product added successfully",
			"id":      result.InsertedID,
		})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// fetchProductsFromDB retrieves products from MongoDB
// fetchProductsFromDB retrieves products from MongoDB
func fetchProductsFromDB() ([]map[string]interface{}, error) {
	collection := client.Database("ecommerce").Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []map[string]interface{}
	if err = cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	return products, nil
}

// logAllProducts fetches and logs all products from the database
func logAllProducts() {
	products, err := fetchProductsFromDB()
	if err != nil {
		log.Printf("Error fetching products: %v", err)
		return
	}

	log.Printf("=== All Products (%d) ===", len(products))
	for i, product := range products {
		// Convert the product to JSON for pretty printing
		productJSON, err := json.MarshalIndent(product, "", "  ")
		if err != nil {
			log.Printf("Error marshaling product data: %v", err)
			continue
		}
		
		log.Printf("Product %d:\n%s", i+1, string(productJSON))
	}
	log.Println("=======================")
}
