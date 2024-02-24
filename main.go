package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Timer struct to represent timer values
type Timer struct {
    Time string `json:"time"`
}

// MongoDB configuration
var collectionName = "timers"
var collection *mongo.Collection
var dbName = "timerdb"

func main() {

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

	password := os.Getenv("PASSWORD")

    // Check if the environment variable is set
    if password == "" {
        fmt.Println("PASSWORD environment variable is not set.")
    }
    
    // Connect to MongoDB
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb+srv://lkalinin:"+ password + "@cluster0.dajck5y.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0").SetServerAPIOptions(serverAPI)
	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
	  panic(err)
	}
	defer func() {
	  if err = client.Disconnect(context.TODO()); err != nil {
		panic(err)
	  }
	}()
	// Send a ping to confirm a successful connection
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
	  panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
  

    // Set up collection
    collection = client.Database(dbName).Collection(collectionName)

    // Define HTTP routes
    http.HandleFunc("/timers", getTimers)
    http.HandleFunc("/timer", addTimer)

    // Enable CORS
    // Middleware function to set CORS headers
    setCORSHeaders := func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Access-Control-Allow-Origin", "*")
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

            if r.Method == "OPTIONS" {
                // Preflight request, respond with success
                w.WriteHeader(http.StatusOK)
                return
            }

            next.ServeHTTP(w, r)
        })
    }

    // Start server
    log.Println("Server listening on :8080")
    log.Fatal(http.ListenAndServe(":"+port, setCORSHeaders(http.DefaultServeMux)))
}

// getTimers retrieves all timer values from the database
func getTimers(w http.ResponseWriter, r *http.Request) {
    var timers []Timer

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    cursor, err := collection.Find(ctx, bson.D{})
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer cursor.Close(ctx)

    for cursor.Next(ctx) {
        var timer Timer
        if err := cursor.Decode(&timer); err != nil {
            log.Println(err)
            continue
        }
        timers = append(timers, timer)
    }

    if err := cursor.Err(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(timers)
}

// addTimer adds a new timer value to the database
func addTimer(w http.ResponseWriter, r *http.Request) {
    var timer Timer

    err := json.NewDecoder(r.Body).Decode(&timer)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    _, err = collection.InsertOne(ctx, timer)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
}
