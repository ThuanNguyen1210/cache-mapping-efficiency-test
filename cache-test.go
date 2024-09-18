package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/redis/go-redis/v9"
)

const MAX_ENTRIES = 1000000
const MAX_VALUES = 12000000
const HSET_BUCKET_SIZE = 500

func main() {

	cacheOption := "redis-hset"

	ctx := context.Background()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Flush all data in Redis
	if err := redisClient.FlushAll(ctx).Err(); err != nil {
		fmt.Println("Error flushing Redis database:", err)
		return
	}

	// Initialize a new pipeline
	redisPipeline := redisClient.Pipeline()

	for i := 1; i <= MAX_ENTRIES; i++ {
		value := fmt.Sprintf("%d", rand.IntN(MAX_VALUES))

		switch cacheOption {
		case "redis-getset":
			redisPipeline.Set(ctx, fmt.Sprintf("%d", i), value, 0)
		case "redis-hset":
			bucket := int(i / HSET_BUCKET_SIZE)
			redisPipeline.HSet(ctx, string(bucket), i, value)
		}

		if i%(MAX_ENTRIES/10) == 0 {
			if cacheOption == "redis-getset" || cacheOption == "redis-hset" {
				if _, err := redisPipeline.Exec(ctx); err != nil {
					fmt.Println("Error executing pipeline:", err)
					return
				}
				redisPipeline = redisClient.Pipeline()
			}
		}
	}

	// Execute any remaining commands in the pipeline
	if _, err := redisPipeline.Exec(ctx); err != nil {
		fmt.Println("Error executing final pipeline:", err)
		return
	}

	// Fetch memory usage information
	info, err := redisClient.Info(ctx, "memory").Result()
	if err != nil {
		fmt.Println("Error getting INFO command result:", err)
		return
	}

	// Parse the memory usage information for human-readable format
	var usedMemoryHuman string
	for _, line := range strings.Split(info, "\n") {
		if strings.HasPrefix(line, "used_memory_human:") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				usedMemoryHuman = parts[1]
				break
			}
		}
	}

	// Print the human-readable memory usage
	fmt.Printf("Total Redis memory usage: %s\n", usedMemoryHuman)
}
