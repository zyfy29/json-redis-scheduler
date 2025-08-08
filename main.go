package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

func main() {
	log.SetFlags(log.Lmicroseconds)

	jsonPath := os.Getenv("JSON_PATH")
	redisURL := os.Getenv("REDIS_URL")
	keyName := os.Getenv("KEY_NAME")
	triggerTimeStr := os.Getenv("TRIGGER_TIME")
	if jsonPath == "" || redisURL == "" || keyName == "" || triggerTimeStr == "" {
		log.Fatal("Required environment variables are missing")
	}

	jsonData, err := readJSONFile(jsonPath)
	if err != nil {
		log.Fatalf("Failed to read JSON file: %v", err)
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}
	rdb := redis.NewClient(opt)
	defer rdb.Close()
	ctx := context.Background()
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	triggerTime, err := time.Parse(time.RFC3339Nano, triggerTimeStr)
	if err != nil {
		log.Fatalf("Failed to parse TRIGGER_TIME: %v", err)
	}
	log.Printf("Trigger time set to: %s\n", triggerTime.Format(time.RFC3339Nano))

	waitForTriggerTime(triggerTime)

	log.Println("Trigger time reached! Starting JSON to Redis operation...")
	err = rdb.Set(ctx, keyName, jsonData, 0).Err()
	if err != nil {
		log.Fatalf("Failed to set data in Redis: %v", err)
	}

	log.Printf("Successfully stored JSON data from %s to Redis key %s at %s\n",
		jsonPath, keyName, time.Now().Format(time.RFC3339Nano))
}

func waitForTriggerTime(triggerTime time.Time) {
	now := time.Now()

	// 既に過去の時刻の場合は即座に実行
	if triggerTime.Before(now) || triggerTime.Equal(now) {
		log.Printf("Trigger time (%s) is in the past or now. Executing immediately.\n",
			triggerTime.Format(time.RFC3339Nano))
		return
	}

	// 待機時間を計算
	duration := triggerTime.Sub(now)
	log.Printf("Waiting %v until trigger time (%s)...\n", duration.Round(time.Second), triggerTime.Format(time.RFC3339Nano))

	time.Sleep(duration)
	log.Printf("Reached trigger time: %s\n", triggerTime.Format(time.RFC3339Nano))
}

func readJSONFile(filePath string) ([]byte, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	var jsonCheck interface{}
	if err := json.Unmarshal(data, &jsonCheck); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %v", err)
	}

	return data, nil
}
