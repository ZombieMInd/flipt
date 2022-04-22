package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// RedisCache wraps go-redis client in order to implement Cacher
type RedisCache struct {
	c          *redis.Client
	expiration time.Duration
}

// NewInMemoryCache creates a new RedisCache with the provided expiration and expiration events listen
func NewRedisCache(expiration time.Duration, logger logrus.FieldLogger, conn *redis.Client) *RedisCache {
	pubsub := conn.PSubscribe(context.Background(), "__keyevent@0__:expired")

	go listenExpirationEvents(*pubsub, logger.WithField("cache", "memory"))

	return &RedisCache{
		c:          conn,
		expiration: expiration,
	}
}

func listenExpirationEvents(pubsub redis.PubSub, logger *logrus.Entry) {
	for {
		message, err := pubsub.ReceiveMessage(context.Background())
		if err != nil {
			logger.Debugf("error receiving redis message: %v", err.Error())
		}

		logger.Debugf("event received: %v", message.String())
		cacheEvictionTotal.WithLabelValues("memory").Inc()
		cacheItemCount.WithLabelValues("memory").Dec()
	}
}

func (r *RedisCache) Get(key string) (interface{}, bool) {
	val, err := r.c.Get(context.TODO(), key).Result()
	if err != nil {
		return nil, false
	}

	return val, true
}

func (r *RedisCache) Set(key string, value interface{}) {
	err := r.c.Set(context.TODO(), key, value, r.expiration).Err()
	if err != nil {
		fmt.Printf("%v", err)
	}

	cacheItemCount.WithLabelValues("memory").Inc()
}

func (r *RedisCache) Delete(key string) {
	r.c.Del(context.TODO(), key)
	cacheItemCount.WithLabelValues("memory").Dec()
}

func (r *RedisCache) Flush() {
	r.c.FlushDBAsync(context.TODO())
	cacheFlushTotal.WithLabelValues("memory").Inc()
	cacheItemCount.WithLabelValues("memory").Set(0)
}
