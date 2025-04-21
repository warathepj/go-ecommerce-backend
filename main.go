package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Product represents the product structure in MongoDB
type Product struct {
	ID          int     `json:"id" bson:"id"`
	Name        string  `json:"name" bson:"name"`
	Description string  `json:"description" bson:"description"`
	Price       float64 `json:"price" bson:"price"`
	ImageUrl    string  `json:"imageUrl" bson:"imageUrl"`
}

var client *mongo.Client
var db *mongo.Database

func initMongoDB() error {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	// Get database instance
	db = client.Database("ecommerce")

	fmt.Println("Successfully connected to MongoDB!")
	return nil
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World")
}

func insertMockProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mockProducts := []Product{
		{ID: 1, Name: "Wireless Mouse", Description: "Ergonomic wireless mouse with long battery life.", Price: 25.99, ImageUrl: "https://placehold.co/300x200/e2e8f0/64748b?text=Wireless+Mouse"},
		{ID: 2, Name: "Mechanical Keyboard", Description: "RGB backlit mechanical keyboard with blue switches.", Price: 79.99, ImageUrl: "https://placehold.co/300x200/e2e8f0/64748b?text=Mechanical+Keyboard"},
		{ID: 3, Name: "USB-C Hub", Description: "7-in-1 USB-C hub with HDMI, SD card reader, and USB 3.0 ports.", Price: 35.50, ImageUrl: "https://placehold.co/300x200/e2e8f0/64748b?text=USB-C+Hub"},
		{ID: 4, Name: "Laptop Stand", Description: "Adjustable aluminum laptop stand for better ergonomics.", Price: 22.00, ImageUrl: "https://placehold.co/300x200/e2e8f0/64748b?text=Laptop+Stand"},
		{ID: 5, Name: "Webcam 1080p", Description: "Full HD 1080p webcam with built-in microphone.", Price: 45.99, ImageUrl: "https://placehold.co/300x200/e2e8f0/64748b?text=Webcam"},
		{ID: 6, Name: "Bluetooth Speaker", Description: "Portable waterproof Bluetooth speaker with rich bass.", Price: 55.00, ImageUrl: "https://placehold.co/300x200/e2e8f0/64748b?text=Bluetooth+Speaker"},
	}

	collection := db.Collection("products")

	// Convert the slice of products to a slice of interface{}
	var documents []interface{}
	for _, product := range mockProducts {
		documents = append(documents, product)
	}

	// Insert the documents
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.InsertMany(ctx, documents)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to insert products: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Products inserted successfully",
		"count":   len(result.InsertedIDs),
	})
}

func getProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	collection := db.Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find all products
	cursor, err := collection.Find(ctx, map[string]interface{}{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch products: %v", err), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	// Decode products
	var products []Product
	if err = cursor.All(ctx, &products); err != nil {
		http.Error(w, fmt.Sprintf("Failed to decode products: %v", err), http.StatusInternalServerError)
		return
	}

	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Return products as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
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

	// Register routes
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/api/products/mock", insertMockProducts)
	http.HandleFunc("/api/products", getProducts)

	// Start server
	fmt.Println("Server starting on port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
