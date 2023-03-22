package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"math/rand"
	"strconv"
	"time"
)

var (
	randInt                    = 6
	resultInRandForFetchFromDb = 2
	secondForRestoreCache      = 10.00
	durationTtlCacheSeconds    = 30
)

type User struct {
	username string
	age      int64
}

var (
	connectUri = "mongodb://user:pass@mongodb:27017/"
	table      = "testing"
	collection = "users"
)

func MakeUser() error {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(connectUri))
	if err != nil {
		return err
	}
	usersCollection := client.Database(table).Collection(collection)

	_, err = usersCollection.InsertMany(context.TODO(), []interface{}{
		User{username: String(15), age: 5},
		User{username: String(15), age: 2},
		User{username: String(15), age: 7},
		User{username: String(15), age: 9},
		User{username: String(15), age: 14},
		User{username: String(15), age: 18},
		User{username: String(15), age: 22},
	})

	if err != nil {
		return err
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	return nil
}

func FetchFirstFiveUser(limit int64) (string, error) {
	ctx := context.Background()

	rdb := redis.NewFailoverClient(&redis.FailoverOptions{
		SentinelAddrs: []string{"redis-sentinel:26379"},
		MasterName:    "mymaster",
	})

	ttl, err := rdb.TTL(ctx, key(limit)).Result()

	if err != nil {
		log.Error(err)
		return fetchFromDB(limit, rdb, ctx)
	}

	log.Error(ttl.Seconds())

	// wrapper
	if ttl.Seconds() < secondForRestoreCache {
		rand.Seed(time.Now().UnixNano())

		if rand.Intn(randInt) == resultInRandForFetchFromDb {
			log.Info("update cache in progress...")
			return fetchFromDB(limit, rdb, ctx)
		}

		return fetchFromCache(limit, rdb, ctx)
	}

	response, err := fetchFromCache(limit, rdb, ctx)

	if rdb.Close() != nil {
		log.Error(err)
	}

	return response, err
}

func key(limit int64) string {
	return "users:" + strconv.Itoa(int(limit))
}

func fetchFromCache(limit int64, rdb *redis.Client, ctx context.Context) (string, error) {
	result, err := rdb.Get(ctx, key(limit)).Result()

	if err != nil {
		log.Error(err)
		return fetchFromDB(limit, rdb, ctx)
	}

	return result, nil
}

func fetchFromDB(limit int64, rdb *redis.Client, ctx context.Context) (string, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(connectUri))
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	usersCollection := client.Database(table).Collection(collection)

	filter := bson.D{}
	opts := options.Find().SetLimit(limit)
	cursor, err := usersCollection.Find(context.TODO(), filter, opts)

	var results []User
	if err = cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	result, err := json.Marshal(results)
	if err != nil {
		fmt.Println(err.Error())

		return "", err
	}

	go storeToCache(limit, rdb, ctx, string(result))

	return string(result), nil
}

func storeToCache(limit int64, rdb *redis.Client, ctx context.Context, result string) {
	duration := time.Duration(durationTtlCacheSeconds) * time.Second

	_, err := rdb.Set(ctx, key(limit), result, duration).Result()
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("cache stored successfully")
}
