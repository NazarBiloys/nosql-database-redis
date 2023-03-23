package service

import (
	"context"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"time"
)

func StoreRandCache() {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	key := String(10)
	duration := time.Duration(30) * time.Minute
	err := rdb.Set(ctx, key, String(10), duration).Err()
	if err != nil {
		log.Error(err)
	}
	log.Info(key)

	defer func(rdb *redis.Client) {
		err := rdb.Close()
		if err != nil {
			log.Error(err)
		}
	}(rdb)
}
