package main

import (
	"WorkQueue/internal/logger"
	"WorkQueue/internal/task"
	"WorkQueue/internal/worker"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var total_jobs_in_queue int64
var jobs_done int = 0
var jobs_failed int = 0

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
	godotenv.Load()
	var wg sync.WaitGroup
	var PORT string = ":" + os.Getenv("PORT_WORKER")

	godotenv.Load()

	rdb := connectRedis()

	n := 3 //Number of goroutines you want to launch

	for i := 0; i < n; i++ {
		wg.Add(1)
		go Run_Worker(rdb, ctx, &wg)
	}

	http.HandleFunc("/metrics", metrics_handler)

	http.ListenAndServe(PORT, nil)

	wg.Wait()

	fmt.Println("All workers finished executing")

}

func metrics_handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		fmt.Fprintf(w, "Only GET request allowed")
		return
	}
	var metrics task.Metrics

	metrics.Total_jobs_in_queue = total_jobs_in_queue
	metrics.Jobs_done = jobs_done
	metrics.Jobs_failed = jobs_failed
	res, err := json.Marshal(metrics)
	if err != nil {
		http.Error(w, "Could not encode metrics", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	fmt.Fprint(w, string(res))

}

func Run_Worker(rdb *redis.Client, ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		res, err := rdb.BLPop(ctx, 0, "task_queue").Result()
		if err != nil {
			log.Println("Error connecting to Redis:", err)
			break
		}
		total_jobs_in_queue, _ = rdb.LLen(ctx, "task_queue").Result()
		var task_to_execute task.Task

		err = json.Unmarshal([]byte(res[1]), &task_to_execute)

		if err != nil {
			log.Fatal("Can not parse data from redis")
			continue
		}

		retries_left := task_to_execute.Retries

		err_worker := worker.Process_Task(task_to_execute)

		if err_worker != nil {
			jobs_failed++
			logger.LogFailure(task_to_execute, err_worker)
			retries_left--
			log.Println("Error processing task:", err_worker, " Putting it back to the queue")
			if retries_left > 0 {
				rdb.RPush(ctx, "task_queue", task_to_execute).Result()
				continue
			} else {
				log.Fatal("Task failed after all retries")
				continue
			}
		} else {
			jobs_done++
			logger.LogSuccess(task_to_execute)
			log.Println("Task done successfully")
		}

	}
}
