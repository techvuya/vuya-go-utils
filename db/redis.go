package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

// CacheClientInterface defines the interface for interacting with a Redis cache.
// It includes methods for setting, getting, and deleting cache data.
type CacheClientInterface interface {
	// Set stores data in Redis with a specific expiration time.
	Set(ctx context.Context, key, data string, expireData time.Duration) error

	// Get retrieves data from Redis using the provided key.
	Get(ctx context.Context, key string) (string, error)

	// Delete removes data from Redis for the provided key.
	Delete(ctx context.Context, key string) error
}

// CacheClient is the Redis client implementation that adheres to the CacheClientInterface.
type CacheClient struct {
	cacheClient *redis.Client
}

// GetRedisClient returns the Redis client used by the CacheClient.
func (a CacheClient) GetRedisClient() *redis.Client {
	return a.cacheClient
}

// Set stores data in Redis with the provided key and expiration duration.
// Returns an error if the operation fails.
func (a CacheClient) Set(ctx context.Context, key, data string, expireData time.Duration) error {
	// _, subSeg := xray.BeginSubsegment(ctx, "query-redis-set")
	// defer subSeg.Close(nil)
	_, err := a.cacheClient.Set(key, data, expireData).Result()
	if err != nil {
		return err
	}
	return nil
}

// Get retrieves the data associated with the given key from Redis.
// Returns the data as a string or an error if the key is not found or another issue occurs.
func (a CacheClient) Get(ctx context.Context, key string) (string, error) {
	// _, subSeg := xray.BeginSubsegment(ctx, "query-redis-get")
	// defer subSeg.Close(nil)
	data, err := a.cacheClient.Get(key).Result()
	if err == redis.Nil {
		return "", ErrQueryNoData
	}
	if err != nil {
		return "", err
	}
	return data, nil
}

// Delete removes the specified key from Redis.
// Returns an error if the key doesn't exist or if another issue occurs.
func (a CacheClient) Delete(ctx context.Context, key string) error {
	// _, subSeg := xray.BeginSubsegment(ctx, "query-redis-delete")
	// defer subSeg.Close(nil)
	_, err := a.cacheClient.Del(key).Result()
	if err == redis.Nil {
		return err
	}
	if err != nil {
		return err
	}
	return nil
}

// Ping checks the connection to Redis
func (a *CacheClient) Ping() error {
	result, err := a.cacheClient.Ping().Result()
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	if result != "PONG" {
		return fmt.Errorf("unexpected ping response: %s", result)
	}
	return nil
}

// Close cleanly shuts down the Redis client
func (a *CacheClient) Close() error {
	return a.cacheClient.Close()
}

// CreateAgCachePoolClient initializes and returns a new Redis client with the specified URL and port.
// It also pings the Redis server to ensure connectivity before returning the client.
// If the connection fails, an error is returned.
func CreateAgCachePoolClient(url, port string) (*CacheClient, error) {
	redisCacheClient := redis.NewClient(&redis.Options{
		Addr:     url + ":" + port,
		PoolSize: 12000,
	})
	result, err := redisCacheClient.Ping().Result()
	if err != nil || result != "PONG" {
		return nil, errors.New("ERROR _CONNECT REDIS")
	}
	CacheClient := &CacheClient{cacheClient: redisCacheClient}
	return CacheClient, nil
}
