package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type OrderItem struct {
	ProductID   int     `json:"productId" bson:"productId"`
	Quantity    int     `json:"quantity" bson:"quantity"`
	PriceAtTime float64 `json:"priceAtTime" bson:"priceAtTime"`
}

type Address struct {
	Street     string `json:"street" bson:"street"`
	City       string `json:"city" bson:"city"`
	State      string `json:"state" bson:"state"`
	PostalCode string `json:"postalCode" bson:"postalCode"`
	Country    string `json:"country" bson:"country"`
}

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "PENDING"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusShipped    OrderStatus = "SHIPPED"
	OrderStatusDelivered  OrderStatus = "DELIVERED"
	OrderStatusCancelled  OrderStatus = "CANCELLED"
)

type Order struct {
	ID              string      `json:"id" bson:"_id"`
	Status          OrderStatus `json:"status" bson:"status"`
	ShippingAddress Address     `json:"shippingAddress" bson:"shippingAddress"`
	Items           []OrderItem `json:"items" bson:"items"`
	Subtotal        float64     `json:"subtotal" bson:"subtotal"`
	Tax             float64     `json:"tax" bson:"tax"`
	Total           float64     `json:"total" bson:"total"`
	CreatedAt       time.Time   `json:"createdAt" bson:"createdAt"`
	UpdatedAt       time.Time   `json:"updatedAt" bson:"updatedAt"`
}

type OrderRequest struct {
	UserDetails struct {
		Name    string `json:"name"`
		Address string `json:"address"`
	} `json:"userDetails"`
	Items    []OrderItem `json:"items"`
	Subtotal float64     `json:"subtotal"`
	Tax      float64     `json:"tax"`
	Total    float64     `json:"total"`
}

func generateOrderID() string {
	// Generate a simple timestamp-based ID
	// In production, you might want to use something more robust like UUID
	return fmt.Sprintf("ORD-%d", time.Now().UnixNano())
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		return
	}

	var orderReq OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&orderReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create order document
	order := Order{
		ID:     generateOrderID(), // You'll need to implement this
		Status: OrderStatus("PENDING"),
		ShippingAddress: Address{
			Street:     orderReq.UserDetails.Address,
			City:       "", // These could be parsed from the address string
			State:      "",
			PostalCode: "",
			Country:    "",
		},
		Items:     orderReq.Items,
		Subtotal:  orderReq.Subtotal,
		Tax:       orderReq.Tax,
		Total:     orderReq.Total,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	collection := db.Collection("orders")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Order created successfully",
		"orderId": order.ID,
	})
}
