package main

import (
	"WorkQueue/internal/task"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

var ctx = context.Background()

func connectRedis() *redis.Client {
	redisURL := os.Getenv("REDIS_URL")
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal("Could not parse Redis URL:", err)
	}
	rdb := redis.NewClient(opt)
	return rdb
}

func main() {
	var PORT string = ":" + os.Getenv("PORT_PRODUCER")

	godotenv.Load()

	rdb = connectRedis()

	http.HandleFunc("/enqueue", post_handler)

	log.Println("Starting the server on port ", PORT)

	err := http.ListenAndServe(PORT, nil)

	if err != nil {
		log.Fatal("error starting the server")
	}

}

func post_handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Println("Only POST request accepted")
		http.Error(w, "Only POST request accepted", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain")

	// Read the request body
	var task task.Task
	err := json.NewDecoder(r.Body).Decode(&task)

	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if task.Type == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if task.Type == "send_email" {
		if task.Payload["to"] == nil || task.Payload["subject"] == nil {
			http.Error(w, "Bad request,pass to and subject fields inside the payload", http.StatusBadRequest)
			return
		}
	}

	b, err := json.Marshal(task)

	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	res1, err := rdb.RPush(ctx, "task_queue", b).Result()

	fmt.Println("Length of queue ", res1)

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Task of type '%s' has been successfully added to the queue", task.Type)

}
